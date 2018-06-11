package tgrpc

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/toukii/goutils"
	"github.com/toukii/jsnm"
)

type Verifier interface {
	Verify(bs []byte)
	VerifyCost(cost int64, costch chan int64)
}

var (
	ns = float64(1e6)
)

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

func (r *Resp) verifyArrLen(js *jsnm.Jsnm, wantLen int, path ...string) (bool, int) {
	arr := js.ArrGet(path...).Arr()
	l := len(arr)
	if l != wantLen {
		return false, l
	}
	return true, l
}

func (r *Resp) VerifyJson(bs []byte) {
	js := jsnm.BytesFmt(bs)
	for ks, wv := range r.Json {
		kss := strings.Split(ks, ",")
		switch kss[len(kss)-1] {
		case "$len":
			wantLen, err := strconv.Atoi(fmt.Sprint(wv))
			if err == nil {
				if oklen, l := r.verifyArrLen(js, wantLen, kss[:len(kss)-1]...); !oklen {
					log.Errorf("%s, want-len:%d, got:%d", ks, wantLen, l)
				}
				continue
			} else {
				log.WithField("path", ks).Errorf("want-len(%+v) is not int!", wv)
			}
		}
		v := js.ArrGet(kss...).RawData().Raw()
		typv := reflect.TypeOf(v)
		typwv := reflect.TypeOf(wv)
		if !reflect.DeepEqual(v, wv) {
			log.WithField("path", ks).Errorf("%+v [%+v] is goten, %+v [%s] is wanted.", v, typv, wv, typwv)
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

func (r *Resp) VerifyCost(cost int64, costch chan int64) {
	if r.Cost == nil {
		return
	}
	if cap(costch) > 1 {
		costch <- cost
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

func summary(key string, ch chan int64, cloz, wait chan bool, n int) {
	sum := 0.0
	size := 0
	max, min := float64(0.0), float64(1e10)
	for {
		select {
		case cost := <-ch:
			fcost := float64(cost)
			sum += fcost
			size++
			if fcost > max {
				max = fcost
			}
			if fcost < min {
				min = fcost
			}
		case <-cloz:
			goto SUMARY
		}
	}

SUMARY:
	if size <= 0 {
		size = 1
	}
	avg := sum / float64(size)
	fmt.Printf("%s size: %d\n avg: %f ms\n max: %f ms\n min: %f ms\n", key, size, avg/ns, max/ns, min/ns)
	wait <- true
}
