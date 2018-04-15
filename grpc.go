package tgrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	jsonpbMarshaler = &jsonpb.Marshaler{}
	log             *logrus.Entry
)

func init() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	log = logrus.NewEntry(logger)
	log.Debug("set log.Level: debug")
}

type tgr struct {
	err           error
	Address       string
	conn          *grpc.ClientConn
	KeepaliveTime time.Duration
	UseExistDesp  bool

	ProtoBasePath  string                              // proto 文件根目录
	IncludeImports string                              // 要执行的方法所在的proto
	sources        map[string]grpcurl.DescriptorSource // 缓存DescriptorSource
}

func (t *tgr) isErr() bool {
	return t.err != nil
}

func (t *tgr) getDescriptorSource(method string) (grpcurl.DescriptorSource, error) {
	if t.isErr() {
		return nil, t.err
	}
	if t.sources == nil {
		t.sources = make(map[string]grpcurl.DescriptorSource)
	}
	if source, ex := t.sources[method]; ex {
		return source, nil
	}
	fileDescriptorSet, err := GetDescriptro(t.ProtoBasePath, method, t.IncludeImports)
	if isErr("get Descriptor", err) {
		t.err = err
		return nil, err
	}

	/*serviceName, err := getServiceName(method)
	if isErr("get ServiceForMethod", err) {
		t.err = err
		return nil, err
	}
	service, err := GetService([]*descriptor.FileDescriptorSet{fileDescriptorSet}, serviceName)
	if isErr("get Service", err) {
		t.err = err
		return nil, err
	}
	fileDescriptorSet, err = SortFileDescriptorSet(service.FileDescriptorSet, service.FileDescriptorProto)
	if isErr("sort FileDescriptorSet", err) {
		t.err = err
		return nil, err
	}*/

	source, err := grpcurl.DescriptorSourceFromFileDescriptorSet(fileDescriptorSet)
	if isErr("grpcurl.DescriptorSource", err) {
		t.err = err
	}
	t.sources[method] = source
	return source, err
}

func SortFileDescriptorSet(fileDescriptorSet *descriptor.FileDescriptorSet, fileDescriptorProto *descriptor.FileDescriptorProto) (*descriptor.FileDescriptorSet, error) {
	// best-effort checks
	names := make(map[string]struct{}, len(fileDescriptorSet.File))
	for _, iFileDescriptorProto := range fileDescriptorSet.File {
		if iFileDescriptorProto.GetName() == "" {
			return nil, fmt.Errorf("no name on FileDescriptorProto")
		}
		if _, ok := names[iFileDescriptorProto.GetName()]; ok {
			return nil, fmt.Errorf("duplicate FileDescriptorProto in FileDescriptorSet: %s", iFileDescriptorProto.GetName())
		}
		names[iFileDescriptorProto.GetName()] = struct{}{}
	}
	if _, ok := names[fileDescriptorProto.GetName()]; !ok {
		return nil, fmt.Errorf("no FileDescriptorProto named %s in FileDescriptorSet with names %v", fileDescriptorProto.GetName(), names)
	}
	newFileDescriptorSet := &descriptor.FileDescriptorSet{}
	for _, iFileDescriptorProto := range fileDescriptorSet.File {
		if iFileDescriptorProto.GetName() != fileDescriptorProto.GetName() {
			newFileDescriptorSet.File = append(newFileDescriptorSet.File, iFileDescriptorProto)
		}
	}
	newFileDescriptorSet.File = append(newFileDescriptorSet.File, fileDescriptorProto)
	return newFileDescriptorSet, nil
}

func (t *tgr) Dial() {
	if t.isErr() {
		return
	}
	log.Debugf("dial tcp:%s ...", t.Address)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	t.conn, t.err = grpcurl.BlockingDial(ctx, "tcp", t.Address, nil, grpc.WithKeepaliveParams(
		keepalive.ClientParameters{
			Time:    t.KeepaliveTime,
			Timeout: t.KeepaliveTime,
		},
	))
	isErr("grpcurl.BlockingDial", t.err)
}

func (t *tgr) Invoke(method string, headers []string, data string) error {
	if t.isErr() {
		return t.err
	}
	source, err := t.getDescriptorSource(method)
	if isErr("get DescriptorSource", err) {
		return err
	}

	methodName, err := getMethod(method)
	if isErr("get Method", err) {
		return err
	}

	err = grpcurl.InvokeRpc(context.Background(),
		source, t.conn, methodName, headers,
		newInvocationEventHandler(), decodeFunc(strings.NewReader(data)))
	isErr("grpcurl.InvokeRpc", err)
	t.err = err
	return err
}

func decodeFunc(reader io.Reader) func() ([]byte, error) {
	decoder := json.NewDecoder(reader)
	return func() ([]byte, error) {
		var rawMessage json.RawMessage
		if err := decoder.Decode(&rawMessage); err != nil {
			return nil, err
		}
		return rawMessage, nil
	}
}

type invocationEventHandler struct {
	err error
}

func newInvocationEventHandler() *invocationEventHandler {
	return &invocationEventHandler{}
}

func (i *invocationEventHandler) OnResolveMethod(*desc.MethodDescriptor) {}

func (i *invocationEventHandler) OnSendHeaders(metadata.MD) {}

func (i *invocationEventHandler) OnReceiveHeaders(metadata.MD) {}

func (i *invocationEventHandler) OnReceiveResponse(message proto.Message) {
	s, err := jsonpbMarshaler.MarshalToString(message)
	if isErr("Marshal", err) {
		return
	}
	log.WithField("OnReceiveResponse", s).Info()
}

func (i *invocationEventHandler) OnReceiveTrailers(s *status.Status, _ metadata.MD) {
	if err := s.Err(); err != nil {
		// TODO(pedge): not great for streaming
		i.err = err
		log.WithField("OnReceiveTrailers", err).Error()
		// printed by returning the error in handler
		//i.println(err.Error())
	}
}

func isErr(source string, err error) bool {
	if err != nil {
		log.Infof("%s err:%s", source, err)
		return true
	}
	return false
}
