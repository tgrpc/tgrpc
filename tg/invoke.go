package main

import (
	"github.com/tgrpc/tgrpc"
)

type Invoke struct {
	Method  string   `toml:"method"`
	Headers []string `toml:"headers"`
	Data    string   `toml:"data"`
}

type Rpc struct {
	Tgr     *tgrpc.Tgrpc `toml:"tgrpc"`
	Invokes []*Invoke    `toml:"invokes"`
}
