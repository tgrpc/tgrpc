package main

import (
	"time"

	"github.com/BurntSushi/toml"
	"github.com/tgrpc/tgrpc"
	"github.com/toukii/bytes"
	"github.com/toukii/goutils"
)

type Tgrpc struct {
	Service  map[string]*tgrpc.Tgrpc `toml:"service"`
	Invokes  []*tgrpc.Invoke         `toml:"invokes"`
	LogLevel string                  `toml:"log_level"`
}

func Setup() *Tgrpc {
	tgr := new(Tgrpc)
	_, err := toml.DecodeFile(conf, tgr)
	if err != nil {
		log.Fatal(err)
	}
	return tgr
}

func initrpc() {
	invoke1 := &tgrpc.Invoke{
		GrpcService: "Greeter",
		Method:      "helloworld.Greeter/SayHello",
		Headers:     []string{"customerId:123", "region:UK"},
		Data:        `{"name":"tgrpc-tg1"}`,
		N:           2,
		Interval:    &tgrpc.Ms{time.Millisecond * 200},
		Resp: &tgrpc.Resp{
			Cost: &tgrpc.Ms{time.Millisecond * 500},
			Json: map[string]interface{}{
				`message`: "Hello tgrpc-tg1",
			},
		},
	}
	invoke2 := &tgrpc.Invoke{
		GrpcService: "LangService",
		Method:      "helloworld.LangService/List",
		Headers:     []string{"customerId:123", "region:UK"},
		Data:        `{}`,
		N:           2,
		Interval:    &tgrpc.Ms{time.Millisecond * 200},
		Resp: &tgrpc.Resp{
			Cost: &tgrpc.Ms{time.Millisecond * 500},
			Json: map[string]interface{}{
				`langs,0,birthday`:           1136185445.0,
				`totalCount`:                 999.0,
				`langs,1,versions,0,id`:      700.0,
				`langs,1,versions,0,vi32s,0`: 111.0,
				`langs,1,versions,0,vi32s`:   []float64{111.0, 222.0},
				`langs,1,versions,0,vi64s,0`: 111.0,
				`langs,1,versions,0,vi64s`:   []float64{111.0, 222.0},
				`langs,1,versions,0,vf64s,0`: 111.0,
				`langs,1,versions,0,vf64s`:   []float64{111.0, 222.0},
				`langs,1,versions,0,vstrs,0`: "str1",
				`langs,1,versions,0,vstrs`:   []string{"str1", "str2"},
			},
		},
		Next: invoke1,
	}

	invokes := []*tgrpc.Invoke{
		invoke1,
		invoke2,
		invoke2,
	}

	tgrs := &Tgrpc{
		Service: map[string]*tgrpc.Tgrpc{
			"Greeter": &tgrpc.Tgrpc{
				Address:        "localhost:2080",
				KeepaliveTime:  &tgrpc.Duration{time.Second * 100},
				ReuseDesc:      true,
				ProtoBasePath:  "$GOPATH/src/github.com/tgrpc/ngrpc",
				IncludeImports: "helloworld/helloworld.proto",
			},
			"LangService": &tgrpc.Tgrpc{
				Address:        "localhost:2080",
				KeepaliveTime:  &tgrpc.Duration{time.Second * 100},
				ReuseDesc:      true,
				ProtoBasePath:  "$GOPATH/src/github.com/tgrpc/ngrpc",
				IncludeImports: "helloworld/lang.proto",
			},
		},
		Invokes:  invokes,
		LogLevel: "debug",
	}

	wr := bytes.NewWriter(make([]byte, 0, 256))
	err := toml.NewEncoder(wr).Encode(tgrs)
	log.Infof("encode:\n%s\nerr: %+v", wr.Bytes(), err)
	err = goutils.SafeWriteFile("tgrpc.toml", wr.Bytes())
	if err != nil {
		log.Error(err)
	}
}
