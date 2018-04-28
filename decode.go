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
	if raw == "" {
		return ""
	}
	ret := raw
	js := jsnm.BytesFmt(prebs)
	allpaths := subDecode(raw, true)
	for _, it := range allpaths {
		paths := strings.Split(it, ",")
		val := js.ArrGet(paths...).RawData().Raw()
		if val == nil {
			continue
		}
		vv := value(val)
		if vv != "" {
			ret = strings.Replace(ret, fmt.Sprintf(`@%s`, it), fmt.Sprintf(`%s`, vv), -1)
		}
	}
	return ret
}

func subDecode(raw interface{}, first bool) []string {
	if first {
		var vals interface{}
		err := json.Unmarshal([]byte(fmt.Sprint(raw)), &vals)
		if err != nil {
			log.Errorf("%+v, err:%+v", raw, err)
			return nil
		}
		return decodeMap(vals)
	}
	switch typ := raw.(type) {
	case string:
		if retlet := getLetterStr([]byte(fmt.Sprint(raw))); retlet != "" {
			return []string{retlet}
		}
	case []interface{}:
		return decodeSlice(raw)
	case map[string]interface{}:
		return decodeMap(raw)
	default:
		log.Debugf("%+v decode unsupported!", typ)
	}
	return nil
}

func decodeMap(raw interface{}) []string {
	vs, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	ret := make([]string, 0, 1)
	for _, subit := range vs {
		if subret := subDecode(subit, false); len(subret) > 0 {
			ret = append(ret, subret...)
		}
	}
	return ret
}

func decodeSlice(raw interface{}) []string {
	vs, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	ret := make([]string, 0, 1)
	for _, subit := range vs {
		if subret := subDecode(subit, false); len(subret) > 0 {
			ret = append(ret, subret...)
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
		log.Infof("%+v value unsupported!", typ)
	}
	return ""
}

// 返回符合jsnm ArrGet的路径，以@开头,以#结尾
func getLetterStr(bs []byte) string {
	if bs[0] != at {
		return ""
	}
	rs := bytes.Runes(bs)
	size := len(rs)
	for i := 1; i < size; i++ {
		if unicode.IsLetter(rs[i]) || unicode.IsNumber(rs[i]) || rs[i] == comma || rs[i] == dblquot {
			continue
		}
		return goutils.ToString(bs[1:i])
	}
	return goutils.ToString(bs[1:])
}
