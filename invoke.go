package tgrpc

import (
	"fmt"
	"sync"
	"time"

	"github.com/toukii/goutils"
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
	i.WaitRet = make(chan bool, 1)
	i.Clozch = make(chan bool, 1)

	if i.N > 1 && i.Resp != nil {
		i.Costch = make(chan int64, 10)
		go summary(i.Method, i.Costch, i.Clozch, i.WaitRet, i.N)
	} else {
		i.WaitRet <- true
	}
}

type Ms struct {
	time.Duration
}

func (d *Ms) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(goutils.ToString(text))
	return err
}

func (d *Ms) MarshalText() ([]byte, error) {
	return goutils.ToByte(fmt.Sprintf("%dms", int64(d.Nanoseconds()/1e6))), nil
}
