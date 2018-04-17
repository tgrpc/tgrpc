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

	tg := Setup()
	if len(tg) <= 0 {
		log.Errorf("config is nil")
	}
	for k, tgr := range tg {
		log.WithField("tgrpc", k).Infof("exced:%t", tgr.Exced)
		if !tgr.Exced {
			continue
		}
		tgr.Tgr.Dial()
		for _, inv := range tgr.Invokes {
			if !inv.Exced {
				continue
			}
			tgr.Tgr.Invoke(inv.Method, inv.Headers, inv.Data)
		}
	}
}

func initSetup() {
	rpc := Rpc{}
	rpc.Tgr = &tgrpc.Tgrpc{
		Address:        "localhost:80",
		KeepaliveTime:  &tgrpc.Duration{time.Second * 100},
		ReuseDesc:      true,
		ProtoBasePath:  "/Users/toukii/PATH/GOPATH/ezbuy/goflow/src/github.com/toukii/ngrpc",
		IncludeImports: "helloworld/helloworld.proto",
	}
	ivk := &Invoke{
		Method:  "helloworld.Greeter/SayHello",
		Headers: []string{"customerId:123", "region:UK"},
		Data:    `{"name":"tgrpc-tg1"}`,
		Exced:   true,
	}
	ivk2 := &Invoke{
		Method:  "helloworld.Greeter/SayHello",
		Headers: []string{"customerId:345", "region:UK"},
		Data:    `{"name":"tgrpc-tg2"}`,
		Exced:   true,
	}
	rpc.Invokes = []*Invoke{ivk, ivk2}
	rpc.Exced = true
	wr := bytes.NewWriter(make([]byte, 0, 256))
	tgrs := TG{
		"rpc1": &rpc,
		"rpc2": &rpc,
	}
	err := toml.NewEncoder(wr).Encode(tgrs)
	log.Infof("encode:\n%s\nerr: %+v", wr.Bytes(), err)
	err = goutils.SafeWriteFile("tgrpc.toml", wr.Bytes())
	if err != nil {
		log.Error(err)
	}
}

func Setup() TG {
	var rs TG
	_, err := toml.DecodeFile(conf, &rs)
	if err != nil {
		log.Panic(err)
	}
	return rs
}
