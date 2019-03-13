package main

import (
	"flag"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tgrpc/tgrpc"
)

var (
	conf    string
	initial bool
	silence bool
	curl    bool

	log *logrus.Entry
)

func init() {
	flag.StringVar(&conf, "c", "tgrpc.toml", "-c tgrpc.toml")
	flag.BoolVar(&initial, "i", false, "-i : init")
	flag.BoolVar(&silence, "s", false, "-s : silence")
	flag.BoolVar(&curl, "C", false, "-C : curl")

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
		initrpc()
		return
	}

	_tgrpc := Setup()
	tgrpc.Silence = silence
	tgrpc.Curl = curl
	tgrpc.SetLog(_tgrpc.LogLevel)
	if _tgrpc.Service == nil {
		log.Errorf("services is nil, all invokes down!")
		return
	}

	invoked := false
	for _, service := range _tgrpc.Service {
		for _, data := range service.Datas {
			invoked = true
			for _, ivk := range _tgrpc.Invokes {
				service.Data = data
				// ivk.Data = data
				tgrpc.Invokes(_tgrpc.Service, ivk)
			}
			// time.Sleep(time.Second * 30)
			time.Sleep(1)
		}
	}

	if invoked {
		return
	}
	for _, ivk := range _tgrpc.Invokes {
		tgrpc.Invokes(_tgrpc.Service, ivk)
	}
}
