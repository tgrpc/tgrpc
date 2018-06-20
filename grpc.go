package tgrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/sirupsen/logrus"
	"github.com/tgrpc/grpcurl"
	"github.com/tgrpc/jdecode"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	indent          = "	"
	jsonpbMarshaler = &jsonpb.Marshaler{Indent: indent}
	log             *logrus.Entry
)

func init() {
	SetLog("debug")
}

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
	sync.Once
	sync.Mutex
	err     error
	conn    *grpc.ClientConn
	sources map[string]grpcurl.DescriptorSource // 缓存DescriptorSource

	Data           string  `toml:"data"`
	Address        string  `toml:"address"`
	KeepaliveTime  *Second `toml:"keepalive"`
	ReuseDesc      bool    `toml:"reuse_desc"`
	ProtoBasePath  string  `toml:"proto_base_path"` // proto 文件根目录
	IncludeImports string  `toml:"include_imports"` // 要执行的方法所在的proto
}

func (t *Tgrpc) isErr() bool {
	return t.err != nil
}

func (t *Tgrpc) getDescriptorSource(method string) (grpcurl.DescriptorSource, error) {
	t.Lock()
	defer t.Unlock()
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
	service, err := GetServiceDescriptor([]*descriptor.FileDescriptorSet{fileDescriptorSet}, serviceName)
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

func (t *Tgrpc) dial() {
	t.Do(func() {
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
	})
}

func (t *Tgrpc) Invoke(ivk *Invoke) error {
	t.dial()
	if t.isErr() {
		return t.err
	}
	source, err := t.getDescriptorSource(ivk.Method)
	if isErr(err) {
		return err
	}

	methodName, err := getMethod(ivk.Method)
	if isErr(err) {
		return err
	}
	if ivk.Next != nil && ivk.Next.preResp == nil {
		ivk.Next.preResp = make(chan []byte, ivk.N)
	}

	var datas []string
	// pre invoke resp
	if ivk.preResp != nil {
		bs := <-ivk.preResp
		if cap(ivk.preResp) == 1 || ivk.N > 1 && len(ivk.preResp) < cap(ivk.preResp)-1 { // 容量不够写，不要再往回放
			ivk.preResp <- bs
		}
		datas, _ = jdecode.Decode(ivk.Data, bs)
	} else {
		if t.Data != "" {
			datas, _ = jdecode.Decode(ivk.Data, []byte(t.Data))
			log.Infof("DecodeData: %+v, %s ==> %+v", ivk.Data, t.Data, datas)
		} else {
			datas = []string{ivk.Data}
		}
	}

	for _, data := range datas {
		if !Silence {
			log.Infof("data: %+v", data)
		}
		err = grpcurl.InvokeRpc(context.Background(),
			source, t.conn, methodName, ivk.Headers,
			newInvocationEventHandler(ivk.Resp, methodName, ivk, ivk.Next), decodeFunc(strings.NewReader(data)))
		isErr(err)
	}
	return nil
}

func Invokes(service map[string]*Tgrpc, ivk *Invoke) {
	if ivk == nil || ivk.N <= 0 {
		return
	}
	sg := sync.WaitGroup{}
	for i := 0; i < ivk.N; i++ {
		sg.Add(1)
		go func(i int) {
			rpc, ok := service[ivk.GrpcService]
			if !ok {
				log.Errorf("service:[%s] is not found!", ivk.GrpcService)
				sg.Done()
				return
			}

			err := rpc.Invoke(ivk)
			if err != nil {
				isErr(err)
				log.Errorf("rpc resp err:%+v", err)
			}
			sg.Done()
			Invokes(service, ivk.Next)
		}(i)
		if ivk.Interval != nil {
			time.Sleep(time.Duration(ivk.Interval.Nanoseconds()))
		}
	}
	sg.Wait()
	if ivk.N > 1 && ivk.Resp != nil {
		ivk.Clozch <- true
		<-ivk.WaitRet
	}
	for _, ivk := range ivk.Then {
		Invokes(service, ivk)
	}
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
