package tgrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/toukii/goutils"
	"github.com/toukii/jsnm"
)

type iv map[string]interface{}

var (
	at      = "@"[0]
	comma   = rune(","[0])
	dblquot = rune(`"`[0])
)

func Decode(raw string, prebs []byte) string {
	var v iv
	err := json.Unmarshal([]byte(raw), &v)
	if isErr(err) {
		return raw
	}
	js := jsnm.BytesFmt(prebs)
	ret := raw
	for _, it := range v {
		switch it.(type) {
		case string:
			_path_ := getLetterStr([]byte(fmt.Sprint(it)))
			if _path_ == "" {
				continue
			}
			paths := strings.Split(_path_, ",")
			val := js.ArrGet(paths...).RawData().Raw()
			if val == nil {
				continue
			}
			vv := value(val)
			if vv != "" {
				log.Debugf("%#v %+v ==> %s", val, val, vv)
				ret = strings.Replace(ret, fmt.Sprintf(`@%s`, _path_), fmt.Sprintf(`%s`, vv), -1)
			} else {
				ret = strings.Replace(ret, fmt.Sprintf(`@%s`, _path_), fmt.Sprintf(`%+v`, val), -1)
			}
		}
	}
	return ret
}

func value(v interface{}) string {
	switch typ := v.(type) {
	case int:
		return fmt.Sprint(v.(int))
	case int32:
		return fmt.Sprint(v.(int32))
	case int64:
		return fmt.Sprint(v.(int64))
	case float32:
		vv := v.(float32)
		return fmt.Sprint(int64(vv))
	case float64:
		vv := v.(float64)
		return fmt.Sprint(int64(vv))
	case string:
		return fmt.Sprint(v)
	default:
		log.Infof("%+v, typ: %+v", v, typ)
	}
	return ""
}

// 返回符合jsnm ArrGet的路径，以@开头,以#结尾
func getLetterStr(bs []byte) string {
	idx := 1
	if bs[0] != at {
		idx = 0
	}
	rs := bytes.Runes(bs)
	size := len(rs)
	for i := idx; i < size; i++ {
		if unicode.IsLetter(rs[i]) || unicode.IsNumber(rs[i]) || rs[i] == comma || rs[i] == dblquot {
			continue
		}
		return goutils.ToString(bs[idx:i])
	}
	return goutils.ToString(bs[idx:])
}
