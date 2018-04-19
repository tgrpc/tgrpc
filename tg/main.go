package main

import (
	"flag"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tgrpc/tgrpc"
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
		log.Errorf("config is nil, init: tg -i")
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
