package main

import (
	"time"

	"github.com/BurntSushi/toml"
	"github.com/tgrpc/tgrpc"
	"github.com/toukii/bytes"
	"github.com/toukii/goutils"
)

type Rpc struct {
	Tgr      *tgrpc.Tgrpc    `toml:"tgrpc"`
	Invokes  []*tgrpc.Invoke `toml:"invokes"`
	Exced    bool            `toml:"exced"` // 是否执行
	LogLevel string          `toml:"log_level"`
}

type TG map[string]*Rpc

func Setup() TG {
	var rs TG
	_, err := toml.DecodeFile(conf, &rs)
	if err != nil {
		log.Error(err)
	}
	return rs
}

func initSetup() {
	rpc := Rpc{
		Tgr: &tgrpc.Tgrpc{
			Address:        "localhost:2080",
			KeepaliveTime:  &tgrpc.Duration{time.Second * 100},
			ReuseDesc:      true,
			ProtoBasePath:  "$GOPATH/src/github.com/tgrpc/ngrpc",
			IncludeImports: "helloworld/helloworld.proto",
		},
		Invokes: []*tgrpc.Invoke{
			&tgrpc.Invoke{
				Method:   "helloworld.Greeter/SayHello",
				Headers:  []string{"customerId:123", "region:UK"},
				Data:     `{"name":"tgrpc-tg1"}`,
				N:        5,
				Interval: &tgrpc.Ms{time.Millisecond * 200},
				Resp: &tgrpc.Resp{
					Cost: &tgrpc.Ms{time.Millisecond * 500},
					Json: map[string]interface{}{
						`message`: "Hello tgrpc-tg1",
					},
				},
			},
		},
		Exced:    true,
		LogLevel: "debug",
	}

	wr := bytes.NewWriter(make([]byte, 0, 256))
	tgrs := TG{
		"Greeter": &rpc,
	}
	err := toml.NewEncoder(wr).Encode(tgrs)
	log.Infof("encode:\n%s\nerr: %+v", wr.Bytes(), err)
	err = goutils.SafeWriteFile("tgrpc.toml", wr.Bytes())
	if err != nil {
		log.Error(err)
	}
}
