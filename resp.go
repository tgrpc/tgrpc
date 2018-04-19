package tgrpc

import (
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/toukii/goutils"
	"github.com/toukii/jsnm"
)

type Verifier interface {
	Verify(bs []byte)
	VerifyCost(cost int64)
}

type Resp struct {
	Cost   *Ms                    `toml:"cost"`
	Body   string                 `toml:"body"`
	Regexp string                 `toml:"regexp"`
	Json   map[string]interface{} `toml:"json"`
}

func (r *Resp) Verify(bs []byte) {
	r.VerifyJson(bs)
	r.VerifyRegexp(bs)
	r.VerifyBody(bs)
}

func (r *Resp) VerifyJson(bs []byte) {
	js := jsnm.BytesFmt(bs)
	for ks, wv := range r.Json {
		kss := strings.Split(ks, ",")
		k := js.ArrGet(kss...).RawData().Raw()
		if !reflect.DeepEqual(k, wv) {
			log.Errorf("response body: <%+v> is goten, <%+v> is wanted.", k, wv)
		}
	}
}

func (r *Resp) VerifyRegexp(bs []byte) {
	if r.Regexp == "" {
		return
	}
	if matched, err := regexp.Match(r.Regexp, bs); isErr(err) || !matched {
		log.Errorf("response body is: %s, not wanted regexp: %s", goutils.ToString(bs), r.Regexp)
	}
}

func (r *Resp) VerifyBody(bs []byte) {
	return
	if !strings.EqualFold(r.Body, goutils.ToString(bs)) {
		log.Errorf("response body is: %s, not wanted: %s", goutils.ToString(bs), r.Body)
	}
}

func (r *Resp) VerifyCost(cost int64) {
	if r.Cost == nil {
		return
	}
	dcost := time.Duration(cost)
	ns := r.Cost.Nanoseconds()
	ms := ns / 1e6
	if cost >= ns {
		log.Errorf("time cost: %+v more than %d ms;", dcost, ms)
	} else if cost > ns*3/4 {
		log.Warnf("time cost: %+v nearby %d ms;", dcost, ms)
	} else {
		log.Debugf("time cost: %+v / %d ms;", dcost, ms)
	}
}
