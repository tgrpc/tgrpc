package main

import (
	"flag"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"github.com/tgrpc/tgrpc"
)

var (
	conf string
	log  *logrus.Entry
)

func init() {
	flag.StringVar(&conf, "c", "tgrpc.toml", "-c tgrpc.toml")

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

	tg := Setup()
	log.Infof("%+v", *tg)
	tg.Dial()
	tg.Invoke("helloworld.Greeter/SayHello", nil, `{"name":"tgrpc"}`)
}

func Setup() *tgrpc.Tgrpc {
	tg := new(tgrpc.Tgrpc)
	_, err := toml.DecodeFile(conf, tg)
	if err != nil {
		log.Panic(err)
	}
	return tg
}
