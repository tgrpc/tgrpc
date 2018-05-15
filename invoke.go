package tgrpc

import (
	"sync"
)

type Invoke struct {
	GrpcService string    `toml:"service"`
	Method      string    `toml:"method"`
	Headers     []string  `toml:"headers"`
	Data        string    `toml:"data"`
	N           int       `toml:"n"`
	Interval    *Ms       `toml:"interval"`
	Resp        *Resp     `toml:"resp"`
	Next        *Invoke   `toml:"next"` // next invoke
	Then        []*Invoke `toml:"then"` // then invoke: all the invokes

	preResp chan []byte
	sync.Once
	Costch  chan int64
	Clozch  chan bool
	WaitRet chan bool
}

func (i *Invoke) Init() {
	log.Infof("%s init", i.Method)
	if i.N > 1 && i.Resp != nil {
		i.WaitRet = make(chan bool, 1)
		i.Clozch = make(chan bool, 1)
		i.Costch = make(chan int64, 10)
		go summary(i.Method, i.Costch, i.Clozch, i.WaitRet, i.N)
	} else {
		go func() {
			<-i.Clozch
			i.WaitRet <- true
		}()
	}
}
