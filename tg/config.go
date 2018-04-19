package main

import (
	"github.com/tgrpc/tgrpc"
)

type Rpc struct {
	Tgr      *tgrpc.Tgrpc    `toml:"tgrpc"`
	Invokes  []*tgrpc.Invoke `toml:"invokes"`
	Exced    bool            `toml:"exced"` // 是否执行
	LogLevel string          `toml:"log_level"`
}

type TG map[string]*Rpc
