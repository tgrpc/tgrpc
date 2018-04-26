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
			[2]string{"msg!", "msg"},
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
