package tgrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/sirupsen/logrus"
	"github.com/tgrpc/grpcurl"
	"github.com/toukii/goutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	jsonpbMarshaler = &jsonpb.Marshaler{}
	log             *logrus.Entry
)

func SetLog(logLevel string) {
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Error(err)
		lvl = logrus.DebugLevel
	}
	logger := logrus.New()
	logger.SetLevel(lvl)
	log = logrus.NewEntry(logger)
}

type Tgrpc struct {
	err     error
	conn    *grpc.ClientConn
	sources map[string]grpcurl.DescriptorSource // 缓存DescriptorSource

	Address        string    `toml:"address"`
	KeepaliveTime  *Duration `toml:"keepalive"`
	ReuseDesc      bool      `toml:"reuse_desc"`
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
	return t.err != nil
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
	fileDescriptorSet, err := GetDescriptor(t.ProtoBasePath, method, t.IncludeImports, t.ReuseDesc)
	if isErr(err) {
		t.err = err
		return nil, err
	}

	serviceName, err := getServiceName(method)
	if isErr(err) {
		t.err = err
		return nil, err
	}
	service, err := GetService([]*descriptor.FileDescriptorSet{fileDescriptorSet}, serviceName)
	if isErr(err) {
		t.err = err
		return nil, err
	}
	fileDescriptorSet, err = SortFileDescriptorSet(service.FileDescriptorSet, service.FileDescriptorProto)
	if isErr(err) {
		t.err = err
		return nil, err
	}

	source, err := grpcurl.DescriptorSourceFromFileDescriptorSet(fileDescriptorSet)
	if isErr(err) {
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
	isErr(t.err)
}

func (t *Tgrpc) Invoke(inv *Invoke) error {
	if t.isErr() {
		return t.err
	}
	source, err := t.getDescriptorSource(inv.Method)
	if isErr(err) {
		return err
	}

	methodName, err := getMethod(inv.Method)
	if isErr(err) {
		return err
	}

	err = grpcurl.InvokeRpc(context.Background(),
		source, t.conn, methodName, inv.Headers,
		newInvocationEventHandler(inv.Resp), decodeFunc(strings.NewReader(inv.Data)))
	isErr(err)
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

func isErr(err error) bool {
	if err != nil {
		func_, file_, line_ := Caller(1)
		fails := logrus.Fields{
			"func": func_,
			"file": fmt.Sprintf("%s :%d", file_, line_),
		}
		log.WithFields(fails).Error(err)
		return true
	}
	return false
}
