package tgrpc

import (
	"strings"
	"testing"
)

var (
	tcases = []struct {
		raw string
		bs  []byte
		des string
	}{
		{
			raw: "",
			bs:  nil,
			des: "",
		},
		{
			raw: `{"name":"@msg"}`,
			bs:  []byte(`{"msg":"success!"}`),
			des: `{"name":"success!"}`,
		},
		{
			raw: `{"name":"@msg!"}`,
			bs:  []byte(`{"msg":"success!"}`),
			des: `{"name":"success!!"}`,
		},
		{
			raw: `{"name":"!@msg!"}`,
			bs:  []byte(`{"msg":"success!"}`),
			des: `{"name":"!@msg!"}`,
		},
		{
			raw: `{"name":"@msg"}`,
			bs:  []byte(`{"no-msg":"success!"}`),
			des: `{"name":"@msg"}`,
		},
		{
			raw: `{"name":"@langs,0,name"}`,
			bs:  []byte(`{"langs":[{"name":"Golang"}]}`),
			des: `{"name":"Golang"}`,
		},
		// {
		// 	raw: `{"names":["@langs,0,name"]}`,
		// 	bs:  []byte(`{"langs":[{"name":"Golang"}]}`),
		// 	des: `{"name":["Golang"]}`,
		// },
	}
)

func TestDecode(t *testing.T) {
	t.Run("Decode-Nil", func(t *testing.T) {
		des := Decode(tcases[0].raw, tcases[0].bs)
		if !strings.EqualFold(des, tcases[0].des) {
			t.Errorf("decode: %s, want: %s, got: %s", tcases[0].raw, tcases[0].des, des)
		} else {
			log.Debugf("decode, raw: %s bs: %s ==> %s", tcases[0].raw, tcases[0].bs, des)
		}
	})

	t.Run("Decode", func(t *testing.T) {
		size := len(tcases)
		for i := 1; i < size; i++ {
			des := Decode(tcases[i].raw, tcases[i].bs)
			if !strings.EqualFold(des, tcases[i].des) {
				t.Errorf("decode: %s, want: %s, got: %s", tcases[i].raw, tcases[i].des, des)
			} else {
				log.Debugf("decode, raw: %s bs: %s ==> %s", tcases[i].raw, tcases[i].bs, des)
			}
		}
	})

	t.Run("getLetterStr", func(t *testing.T) {
		ts := [][2]string{
			[2]string{"@msg!", "msg"},
			[2]string{"msg!", ""},
			[2]string{"!msg!", ""},
			[2]string{"@msg,0", "msg,0"},
			[2]string{"@msg,0,count", "msg,0,count"},
			[2]string{`@msg,0,"1",count`, `msg,0,"1",count`},
			[2]string{`@langs,0,name`, `langs,0,name`},
			[2]string{`@@langs,0,name`, ``},
		}
		for _, it := range ts {
			str := getLetterStr([]byte(it[0]))
			if str != it[1] {
				t.Errorf("%s ==> %s, but: %s", it[0], it[1], str)
			}
		}
	})

}

func TestValue(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		ts := []struct {
			i interface{}
			v string
		}{
			{
				// i: int64(25580228382294197),
				i: 25580228382294197,
				v: "25580228382294197",
			},
			{
				i: 1,
				v: "1",
			},
			{
				i: 10000,
				v: "10000",
			},
		}

		for _, it := range ts {
			vv := value(it.i)
			if !strings.EqualFold(vv, it.v) {
				t.Errorf("%+v ==> %s, but: %s", it.i, it.v, vv)
			}
		}
	})
}

func TestSubDecode(t *testing.T) {
	t.Run("SubDecode", func(t *testing.T) {
		ts := []struct {
			i interface{}
			v []string
		}{
			{
				i: `{"name":"Golang"}`,
				v: nil,
			},
			{
				i: `{"name":"@Golang"}`,
				v: []string{"Golang"},
			},
			{
				i: `{"name":"@Golang!"}`,
				v: []string{"Golang"},
			},
			{
				i: `{"name":["@Golang!","@Golang!"]}`,
				v: []string{"Golang", "Golang"},
			},
			{
				i: `{"name":["@Golang!","@Golang!"],"version":{"prev0":{"val":"@0,name"}}}`,
				v: []string{"Golang", "Golang", "0,name"},
			},
			{
				i: `{"name":["@Golang!","@Golang!"],"version":{"2017":{"prev":["@0,name","@1,name","@2,name"]}}}`,
				v: []string{"Golang", "Golang", "0,name", "1,name", "2,name"},
			},
			{
				i: `{"name":["@Golang!","@Golang!"],"version":[["@0,name","@1,name","@2,name"]]}`,
				v: []string{"Golang", "Golang", "0,name", "1,name", "2,name"},
			},
		}

		for _, it := range ts {
			vv := subDecode(it.i, true)
			if len(vv) <= 0 && len(it.v) <= 0 {
				continue
			}
			if !sliceEqual(it.v, vv) {
				t.Errorf("%+v ==> %s, but: %s", it.i, it.v, vv)
			}
		}
	})
}

// slice 的值和个数相等
func sliceEqual(sl1, sl2 []string) bool {
	s1, s2 := len(sl1), len(sl2)
	if s1 != s2 {
		return false
	}
	type tempty struct{}
	var empty tempty
	m := make(map[string]tempty, s1)
	for _, it := range sl1 {
		m[it] = empty
	}
	var ex bool
	for _, it := range sl2 {
		if _, ex = m[it]; !ex {
			return false
		}
	}
	return true
}
