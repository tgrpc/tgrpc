package tgrpc

import (
	"testing"
	"time"
)

func TestInvokeGRpcGreeter(t *testing.T) {
	tg := &Tgrpc{
		Address:        "localhost:80",
		KeepaliveTime:  Duration{time.Second * 100},
		ProtoBasePath:  ".",
		IncludeImports: "helloworld/helloworld.proto",
	}
	tg.Dial()
	tg.Invoke("helloworld.Greeter/SayHello", nil, `{"name":"tgrpc"}`)
	tg.Invoke("helloworld.Greeter/SayHello", nil, `{"name":"tgrpc"}`)
}
