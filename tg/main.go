package main

import (
	"flag"

	"github.com/sirupsen/logrus"
	"github.com/tgrpc/tgrpc"
)

var (
	conf    string
	initial bool
	silence bool

	log *logrus.Entry
)

func init() {
	flag.StringVar(&conf, "c", "tgrpc.toml", "-c tgrpc.toml")
	flag.BoolVar(&initial, "i", false, "-i : init")
	flag.BoolVar(&silence, "s", false, "-s : silence")

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
	tgrpc.SetLog(_tgrpc.LogLevel)
	if _tgrpc.Service == nil {
		log.Errorf("services is nil, all invokes down!")
		return
	}
	for _, ivk := range _tgrpc.Invokes {
		tgrpc.Invokes(_tgrpc.Service, ivk)
	}
}
