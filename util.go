package tgrpc

import (
	"fmt"
	"time"

	"github.com/toukii/goutils"
)

type Ms struct {
	time.Duration
}

func (m *Ms) UnmarshalText(text []byte) error {
	var err error
	m.Duration, err = time.ParseDuration(goutils.ToString(text))
	return err
}

func (m *Ms) MarshalText() ([]byte, error) {
	return goutils.ToByte(fmt.Sprintf("%dms", int64(m.Nanoseconds()/1e6))), nil
}

type Second struct {
	time.Duration
}

func (s *Second) UnmarshalText(text []byte) error {
	var err error
	s.Duration, err = time.ParseDuration(goutils.ToString(text))
	return err
}

func (s *Second) MarshalText() ([]byte, error) {
	return goutils.ToByte(fmt.Sprintf("%ds", int64(s.Seconds()))), nil
}
