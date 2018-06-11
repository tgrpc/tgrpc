package tgrpc

import (
	"reflect"
	"strings"
	"testing"
)

type testcase struct {
	raw string
	bs  []byte
	des []string
}

var (
	tcases = []testcase{
		{
			raw: "",
			bs:  nil,
			des: []string{""},
		},
		{
			raw: `{"name":"@msg"}`,
			bs:  []byte(`{"msg":"success!"}`),
			des: []string{`{"name":"success!"}`},
		},
		{
			raw: `{"name":"@msg!@msg"}`,
			bs:  []byte(`{"msg":"success!"}`),
			des: []string{`{"name":"success!!@msg"}`},
		},
		{
			raw: `{"name":"@msg!"}`,
			bs:  []byte(`{"msg":"success!"}`),
			des: []string{`{"name":"success!!"}`},
		},
		{
			raw: `{"name":"!@msg!"}`,
			bs:  []byte(`{"msg":"success!"}`),
			des: []string{`{"name":"!@msg!"}`},
		},
		{
			raw: `{"name":"@msg"}`,
			bs:  []byte(`{"no-msg":"success!"}`),
			des: []string{`{"name":"@msg"}`},
		},
		{
			raw: `{"name":"@langs,0,name"}`,
			bs:  []byte(`{"langs":[{"name":"Golang"}]}`),
			des: []string{`{"name":"Golang"}`},
		},
		{
			raw: `{"names":["@langs,0,name"]}`,
			bs:  []byte(`{"langs":[{"name":"Golang"}]}`),
			des: []string{`{"names":["Golang"]}`},
		},
		{
			raw: `{"vals":["@vals,0,i1"]}`,
			bs:  []byte(`{"vals":[{"i1":100}]}`),
			des: []string{`{"vals":[100]}`},
		},
		{
			raw: `{"names":["@Golang!","@Golang!"],"version":[["@versions,0,name","@versions,1,desp","@versions,2,version","@versions,3,version"]]}`,
			bs:  []byte(`{"Golang":"go1.0","versions":[{"name":"v1.0"},{"desp":"desp v1.0"},{"version":1.0},{"version":1.2}]}`),
			des: []string{`{"names":["go1.0!","go1.0!"],"version":[["v1.0","desp v1.0",1,1.2]]}`}, // 1.0 ==> 1
		},
	}
)

func TestDecode(t *testing.T) {
	t.Run("Decode-Nil", func(t *testing.T) {
		des, _ := Decode(tcases[0].raw, tcases[0].bs)
		if !reflect.DeepEqual(des, tcases[0].des) {
			// if !strings.EqualFold(des, tcases[0].des) {
			t.Errorf("decode: %s, want: %s, got: %s", tcases[0].raw, tcases[0].des, des)
		} else {
			log.Debugf("decode, raw: %s bs: %s ==> %s", tcases[0].raw, tcases[0].bs, des)
		}
	})

	t.Run("Decode", func(t *testing.T) {
		size := len(tcases)
		for i := 1; i < size; i++ {
			des, _ := Decode(tcases[i].raw, tcases[i].bs)
			if !reflect.DeepEqual(des, tcases[i].des) {
				// if !strings.EqualFold(des, tcases[i].des) {
				t.Errorf("decode: %s, want: %s, got: %s", tcases[i].raw, tcases[i].des, des)
			} else {
				log.Debugf("decode, raw: %s bs: %s ==> %s", tcases[i].raw, tcases[i].bs, des)
			}
		}
	})

	t.Run("Decode $range", func(t *testing.T) {
		tcases := []testcase{
			{
				raw: `{"vals":["@vals,$range"]}`,
				bs:  []byte(`{"vals":[{"i1":100},{"i2":101}]}`),
				des: []string{`{"vals":[{"i1":100}]}`, `{"vals":[{"i2":101}]}`},
			},
			{
				raw: `{"name":"@$range"}`,
				bs:  []byte(`["1","2","3"]`),
				des: []string{`{"name":"1"}`, `{"name":"2"}`, `{"name":"3"}`},
			},
			{
				raw: `{"val":"@vals,$range"}`,
				bs:  []byte(`{"vals":[1,2,3]}`),
				des: []string{`{"val":1}`, `{"val":2}`, `{"val":3}`},
			},
			{
				raw: `{"val":"@vals,$range","val2":"@vals,$range"}`,
				bs:  []byte(`{"vals":[1,2,3]}`),
				des: []string{`{"val":1,"val2":"@vals,$range"}`, `{"val":2,"val2":"@vals,$range"}`, `{"val":3,"val2":"@vals,$range"}`},
			},
		}
		size := len(tcases)
		for i := 0; i < size; i++ {
			des, _ := Decode(tcases[i].raw, tcases[i].bs)
			if !reflect.DeepEqual(des, tcases[i].des) {
				// if !strings.EqualFold(des, tcases[i].des) {
				t.Errorf("decode: %s, want: %s, got: %s", tcases[i].raw, tcases[i].des, des)
			} else {
				for j, it := range des {
					log.Debugf("%d decode, raw: %s bs: %s ==> %s", j, tcases[i].raw, tcases[i].bs, it)
				}
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
			vv, _ := value(it.i)
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
