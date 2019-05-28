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
	_Costch  chan int64 `toml:"omitempty"`
	_Clozch  chan bool  `toml:"omitempty"`
	_WaitRet chan bool  `toml:"omitempty"`
}

func (i *Invoke) Init() {
	if i.N > 1 && i.Resp != nil {
		i._WaitRet = make(chan bool, 1)
		i._Clozch = make(chan bool, 1)
		i._Costch = make(chan int64, 10)
		go summary(i.Method, i._Costch, i._Clozch, i._WaitRet, i.N)
	}
}

type InvokeDate struct {
	Data string
}
