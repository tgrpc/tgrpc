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
	"github.com/toukii/goutils"
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

type Tgrpc struct {
	err     error
	conn    *grpc.ClientConn
	sources map[string]grpcurl.DescriptorSource // 缓存DescriptorSource

	Address        string    `toml:"address"`
	KeepaliveTime  *Duration `toml:"keepalive"`
	ReuseDesp      bool      `toml:"reuse_desp"`
	ProtoBasePath  string    `toml:"proto_base_path"` // proto 文件根目录
	IncludeImports string    `toml:"include_imports"` // 要执行的方法所在的proto
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(goutils.ToString(text))
	return err
}

func (d *Duration) MarshalText() ([]byte, error) {
	return goutils.ToByte(fmt.Sprintf("%ds", int64(d.Seconds()))), nil
}

func (t *Tgrpc) isErr() bool {
	ret := t.err != nil
	if ret {
		log.Error(t.err)
	}
	return ret
}

func (t *Tgrpc) getDescriptorSource(method string) (grpcurl.DescriptorSource, error) {
	if t.isErr() {
		return nil, t.err
	}
	if t.sources == nil {
		t.sources = make(map[string]grpcurl.DescriptorSource)
	}
	if source, ex := t.sources[method]; ex {
		return source, nil
	}
	fileDescriptorSet, err := GetDescriptro(t.ProtoBasePath, method, t.IncludeImports, t.ReuseDesp)
	if isErr("get Descriptor", err) {
		t.err = err
		return nil, err
	}

	serviceName, err := getServiceName(method)
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
	}

	source, err := grpcurl.DescriptorSourceFromFileDescriptorSet(fileDescriptorSet)
	if isErr("grpcurl.DescriptorSource", err) {
		t.err = err
	}
	t.sources[method] = source
	return source, err
}

func (t *Tgrpc) Dial() {
	if t.isErr() {
		return
	}
	log.Debugf("dial tcp:%s ...", t.Address)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	t.conn, t.err = grpcurl.BlockingDial(ctx, "tcp", t.Address, nil, grpc.WithKeepaliveParams(
		keepalive.ClientParameters{
			Time:    t.KeepaliveTime.Duration,
			Timeout: t.KeepaliveTime.Duration,
		},
	))
	isErr("grpcurl.BlockingDial", t.err)
}

func (t *Tgrpc) Invoke(method string, headers []string, data string) error {
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
