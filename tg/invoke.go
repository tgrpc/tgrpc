package main

import (
	"fmt"
	"time"

	"github.com/tgrpc/tgrpc"
	"github.com/toukii/goutils"
)

type Invoke struct {
	Method   string   `toml:"method"`
	Headers  []string `toml:"headers"`
	Data     string   `toml:"data"`
	N        int      `toml:"n"`
	Interval *Ms      `toml:"interval"`
}

type Rpc struct {
	Tgr      *tgrpc.Tgrpc `toml:"tgrpc"`
	Invokes  []*Invoke    `toml:"invokes"`
	Exced    bool         `toml:"exced"` // 是否执行
	LogLevel string       `toml:"log_level"`
}

type TG map[string]*Rpc

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
