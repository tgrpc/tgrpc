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
		initrpc()
		return
	}

	_tgrpc := Setup()
	tgrpc.SetLog(_tgrpc.LogLevel)
	if _tgrpc.Service == nil {
		log.Errorf("services is nil, all invokes down!")
		return
	}
	for _, ivk := range _tgrpc.Invokes {
		Invoke(_tgrpc.Service, ivk)
	}
}

func Invoke(service map[string]*tgrpc.Tgrpc, ivk *tgrpc.Invoke) {
	if ivk == nil || ivk.N <= 0 {
		return
	}
	sg := sync.WaitGroup{}
	for i := 0; i < ivk.N; i++ {
		sg.Add(1)
		go func(i int) {
			defer sg.Done()
			rpc, ok := service[ivk.GrpcService]
			if !ok {
				log.Errorf("service:[%s] is not found!", ivk.GrpcService)
				return
			}

			err := rpc.Invoke(ivk)
			if err != nil {
				log.Errorf("rpc resp err:%+v", err)
			}
			Invoke(service, ivk.Next)
		}(i)
		if ivk.Interval != nil {
			time.Sleep(time.Duration(ivk.Interval.Nanoseconds()))
		}
	}
	sg.Wait()
	for _, ivk := range ivk.Then {
		Invoke(service, ivk)
	}
}
