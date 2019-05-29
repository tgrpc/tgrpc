package tgrpc

import (
	"sync"

	"github.com/tgrpc/jdecode"
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
