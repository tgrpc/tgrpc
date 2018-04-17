package main

import (
	"github.com/tgrpc/tgrpc"
)

type Invoke struct {
	Method  string   `toml:"method"`
	Headers []string `toml:"headers"`
	Data    string   `toml:"data"`
	Exced   bool     `toml:"exced"` // 是否执行
}

type Rpc struct {
	Tgr     *tgrpc.Tgrpc `toml:"tgrpc"`
	Invokes []*Invoke    `toml:"invokes"`
	Exced   bool         `toml:"exced"` // 是否执行
}

type TG map[string]*Rpc
