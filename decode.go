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
			val := js.ArrGet(paths...).RawData()
			if val == nil {
				continue
			}
			ret = strings.Replace(ret, fmt.Sprintf(`"@%s`, _path_), fmt.Sprintf(`"%+v`, val), -1)
		}
	}
	return ret
}

// 返回符合jsnm ArrGet的路径，可以以@开头
func getLetterStr(bs []byte) string {
	idx := 1
	if bs[0] != at {
		idx = 0
	}
	rs := bytes.Runes(bs)
	// size := len(rs)
	// for i := idx; i < size; i++ {
	for i, it := range rs {
		if i < idx || unicode.IsLetter(it) || unicode.IsNumber(it) || it == comma || it == dblquot {
			continue
		}
		// if i <= idx {
		// 	return ""
		// }
		return goutils.ToString(bs[idx:i])
	}
	return goutils.ToString(bs[idx:])
}
