package tgrpc

import (
	"testing"
	"time"
)

func init() {
	SetLog("debug")
}

func TestInvokeGRpcGreeter(t *testing.T) {
	tg := &Tgrpc{
		Address:        "localhost:2080",
		KeepaliveTime:  &Duration{time.Second * 100},
		ProtoBasePath:  "$GOPATH/src/github.com/tgrpc/ngrpc",
		IncludeImports: "helloworld/helloworld.proto",
	}
	tg.Invoke(&Invoke{
		Method:   "helloworld.Greeter/SayHello",
		Headers:  nil,
		Data:     `{"name":"tgrpc"}`,
		N:        2,
		Interval: &Ms{time.Millisecond * 100},
		Resp: &Resp{
			Json: map[string]interface{}{
				"message": "Hello tgrpc",
			},
		},
	})
}
