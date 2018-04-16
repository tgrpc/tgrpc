package main

import (
	"flag"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"github.com/tgrpc/tgrpc"

	"github.com/toukii/bytes"
	"github.com/toukii/goutils"
)

var (
	conf    string
	initial bool

	log *logrus.Entry
)

func init() {
	flag.StringVar(&conf, "c", "tgrpc.toml", "-c tgrpc.toml")
	flag.BoolVar(&initial, "i", false, "-i")

	setLog()
}

func setLog() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	log = logrus.NewEntry(logger)
	log.Debug("set log.Level: debug")
}

func main() {
	flag.Parse()
	if initial {
		initSetup()
		return
	}

	tgr := Setup()
	log.Infof("%+v", *tgr)
	tgr.Tgr.Dial()
	for _, inv := range tgr.Invokes {
		tgr.Tgr.Invoke(inv.Method, inv.Headers, inv.Data)
	}
}

func initSetup() {
	rpc := Rpc{}
	rpc.Tgr = &tgrpc.Tgrpc{
		Address:        "localhost:80",
		KeepaliveTime:  &tgrpc.Duration{time.Second * 100},
		ReuseDesp:      true,
		ProtoBasePath:  "/Users/toukii/PATH/GOPATH/ezbuy/goflow/src/github.com/toukii/ngrpc",
		IncludeImports: "helloworld/helloworld.proto",
	}
	ivk := &Invoke{
		Method:  "helloworld.Greeter/SayHello",
		Headers: []string{"customerId:123", "region:UK"},
		Data:    `{"name":"tgrpc-tg1"}`,
	}
	ivk2 := &Invoke{
		Method:  "helloworld.Greeter/SayHello",
		Headers: []string{"customerId:345", "region:UK"},
		Data:    `{"name":"tgrpc-tg2"}`,
	}
	rpc.Invokes = []*Invoke{ivk, ivk2}
	wr := bytes.NewWriter(make([]byte, 0, 256))
	err := toml.NewEncoder(wr).Encode(&rpc)
	log.Infof("encode:\n%s\nerr: %+v", wr.Bytes(), err)
	err = goutils.SafeWriteFile("tgrpc.toml", wr.Bytes())
	if err != nil {
		log.Error(err)
	}
}

func Setup() *Rpc {
	r := new(Rpc)
	_, err := toml.DecodeFile(conf, r)
	if err != nil {
		log.Panic(err)
	}
	return r
}
