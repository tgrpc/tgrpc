package main

import (
	"flag"
	"sync"
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

	setLog("debug")
}

func setLog(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.Error(err)
		lvl = logrus.DebugLevel
	}
	logger := logrus.New()
	logger.SetLevel(lvl)
	log = logrus.NewEntry(logger)
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
		if tgr.Tgr == nil {
			log.Infof("%s.Tgrpc is nil", k)
			continue
		}
		tgrpc.SetLog(tgr.LogLevel)
		tgr.Tgr.Dial()
		for _, inv := range tgr.Invokes {
			sg := sync.WaitGroup{}
			n := inv.N
			sg.Add(n)
			for i := 0; i < n; i++ {
				go func(i int) {
					tgr.Tgr.Invoke(inv)
					sg.Done()
				}(i)
				if inv.Interval != nil {
					time.Sleep(time.Duration(inv.Interval.Nanoseconds()))
				}
			}
			sg.Wait()
		}
	}
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
					Json: map[string]string{
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

func Setup() TG {
	var rs TG
	_, err := toml.DecodeFile(conf, &rs)
	if err != nil {
		log.Panic(err)
	}
	return rs
}
