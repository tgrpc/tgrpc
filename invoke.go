package tgrpc

import (
	"sync"

	"github.com/tgrpc/jdecode"
	"github.com/toukii/goutils"
)

type Invoke struct {
	GrpcService string    `toml:"service" yaml:"service"`
	Method      string    `toml:"method" yaml:"method"`
	Headers     []string  `toml:"headers" yaml:"headers"`
	Data        string    `toml:"data" yaml:"data"`
	N           int       `toml:"n" yaml:"n"`
	Interval    *Ms       `toml:"interval" yaml:"interval"`
	Resp        *Resp     `toml:"resp" yaml:"resp"`
	Next        *Invoke   `toml:"next" yaml:"next"` // next invoke
	Then        []*Invoke `toml:"then" yaml:"then"` // then invoke: all the invokes

	preResp chan []byte
	sync.Once
	_Costch  chan int64 `toml:"omitempty" yaml:"omitempty"`
	_Clozch  chan bool  `toml:"omitempty" yaml:"omitempty"`
	_WaitRet chan bool  `toml:"omitempty" yaml:"omitempty"`
}

func (i *Invoke) Init() {
	if i.N > 1 && i.Resp != nil {
		i._WaitRet = make(chan bool, 1)
		i._Clozch = make(chan bool, 1)
		i._Costch = make(chan int64, 10)
		go summary(i.Method, i._Costch, i._Clozch, i._WaitRet, i.N)
	}
}

func (ivk *Invoke) DecodeData(tgrpcDatas []string) (chan string, chan bool, int) {
	ivkData := make(chan string, 4)
	dataEnd := make(chan bool, 4)
	endCount := 0

	if len(tgrpcDatas) <= 0 {
		endCount++
		ivkData <- ivk.Data
		dataEnd <- true
		return ivkData, dataEnd, endCount
	}

	if ivk.preResp != nil {
		bs := <-ivk.preResp
		if cap(ivk.preResp) == 1 || ivk.N > 1 && len(ivk.preResp) < cap(ivk.preResp)-1 { // 容量不够写，不要再往回放
			ivk.preResp <- bs
		}

		endCount++
		jdecode.DecodeByChan(ivk.Data, bs, ivkData, dataEnd)
		return ivkData, dataEnd, endCount
	}

	for _, data_ := range tgrpcDatas {
		tData := jdecode.DecodeDataFile(data_)
		jdecode.DecodeByChan(ivk.Data, goutils.ToByte(tData), ivkData, dataEnd)
		endCount++
	}

	return ivkData, dataEnd, endCount
}
