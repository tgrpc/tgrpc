package tgrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
	"github.com/tgrpc/desc"
	"github.com/tgrpc/grpcurl"
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

	// Data           string   `toml:"data" yaml:"data"` // 废弃，使用datas
	Datas          []string `toml:"datas" yaml:"datas"`
	Address        string   `toml:"address" yaml:"address"`
	KeepaliveTime  *Second  `toml:"keepalive" yaml:"keepalive"`
	ReuseDesc      bool     `toml:"reuse_desc" yaml:"reuse_desc"`
	ProtoBasePath  string   `toml:"proto_base_path" yaml:"proto_base_path"` // proto 文件根目录
	IncludeImports string   `toml:"include_imports" yaml:"include_imports"` // 要执行的方法所在的proto
	RawDescs       []string `toml:"raw_descs" yaml:"raw_descs"`             // raw desc, []byte copy后的字符串
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

	source, err := desc.GetDescriptorSource(t.ProtoBasePath, method, t.IncludeImports, t.ReuseDesc, t.RawDescs)
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
		log.Debugf("dial %s ...", t.Address)
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

	methodName, err := desc.GetMethod(ivk.Method)
	if isErr(err) {
		return err
	}
	if ivk.Next != nil && ivk.Next.preResp == nil {
		ivk.Next.preResp = make(chan []byte, ivk.N)
	}

	ivkData, dataEnd, endCount := ivk.DecodeData(t.Datas)

	for {
		if len(ivkData) <= 0 && endCount <= 0 {
			break
		}
		select {
		case <-dataEnd:
			endCount--
		case data := <-ivkData:
			if !Silence {
				log.Infof("data: %+v", data)
			}
			if Curl {
				fmt.Println(t.Tocurl(ivk, data))
			}
			err = grpcurl.InvokeRpc(context.Background(),
				source, t.conn, methodName, ivk.Headers,
				newInvocationEventHandler(ivk.Resp, methodName, ivk, ivk.Next), decodeFunc(strings.NewReader(data)))
			isErr(err)
			if ivk.Interval != nil {
				time.Sleep(time.Duration(ivk.Interval.Nanoseconds()))
			}
		}
	}

	return nil
}

func (t *Tgrpc) Tocurl(ivk *Invoke, data string) string {
	buf := bytes.NewBuffer(make([]byte, 0, 2014))
	buf.WriteString("curl ")
	http_ := "http://"
	if false {
		http_ = "https://"
	}
	buf.WriteString(http_)
	buf.WriteString(strings.TrimSuffix(strings.TrimSuffix(t.Address, ":2080"), ":2083"))
	buf.WriteString("/api/")
	buf.WriteString(ivk.Method)
	for _, h := range ivk.Headers {
		buf.WriteString(fmt.Sprintf(" -H '%s'", h))
	}
	buf.WriteString(" -X 'POST' -H 'Content-Type: application/json'")
	// buf.WriteString(` -H 'Accept-Encoding: gzip, deflate' -H 'Accept-Language: zh-CN,zh;q=0.8,en;q=0.6,zh-TW;q=0.4' -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36' -H 'Content-Type: text/plain;charset:utf-8' -H 'Accept: /'`)
	buf.WriteString(fmt.Sprintf(" --data-binary '%s'", data))
	buf.WriteString(" --compressed")

	return buf.String()
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
		ivk._Clozch <- true
		<-ivk._WaitRet
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
