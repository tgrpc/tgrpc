package tgrpc

import (
	"regexp"
	"strings"

	"github.com/toukii/goutils"
	"github.com/toukii/jsnm"
)

type Verifier interface {
	Verify(bs []byte, code int32, cost int64)
}

type Resp struct {
	Code   int32
	Cost   *Ms
	Body   string
	Regexp string
	Json   map[string]string
}

func (r *Resp) Verify(bs []byte, code int32, cost int64) {
	r.VerifyJson(bs)
	r.VerifyRegexp(bs)
	r.VerifyBody(bs)
	r.VerifyCode(code)
	r.VerifyCost(cost)
}

func (r *Resp) VerifyJson(bs []byte) {
	js := jsnm.BytesFmt(bs)
	for ks, wv := range r.Json {
		kss := strings.Split(ks, ",")
		k := js.ArrGet(kss...).RawData().String()
		if k != wv {
			log.Errorf("response body: <%s> is goten, <%s> is wanted.", k, wv)
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

func (r *Resp) VerifyCode(code int32) {
	if r.Code != code {
		log.Errorf("error code::%d gotten, %d wanted", code, r.Code)
	}
}

func (r *Resp) VerifyCost(cost int64) {
	if r.Cost == nil {
		return
	}
	rcost := r.Cost.Nanoseconds() / 1e6
	if cost > rcost {
		log.Errorf("time cost: %d ms more than %d ms;", cost, rcost)
	} else if cost > rcost*3/4 {
		log.Warnf("time cost: %d ms near by %d ms;", cost, rcost)
	} else {
		log.Infof("time cost: %d ms / %d ms;", cost, rcost)
	}
}
