package tgrpc

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/golang/protobuf/proto"
	protobuf "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

func ParseStr2Bytes(str string) []byte {
	s := strings.Split(str, ",")
	return Parse2Bytes(s)
}

func Parse2Bytes(strs []string) []byte {
	bs := make([]byte, 0, len(strs))
	for _, it := range strs {
		if it == "" {
			continue
		}
		it = strings.TrimSpace(it)
		bs = append(bs, Parse2Byte(it))
	}
	return bs
}

func Parse2Byte(v string) byte {
	return s2i(v[2])<<4 + s2i(v[3])
}

func s2i(s byte) byte {
	if s <= 57 {
		return s - 48
	}
	// if s >= 97 {
	// }
	return s - 97 + 10
}

// extractFile extracts a FileDescriptorProto from a gzip'd buffer.
func ExtractFile(gz []byte) (*protobuf.FileDescriptorProto, error) {
	r, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, fmt.Errorf("failed to open gzip reader: %v", err)
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to uncompress descriptor: %v", err)
	}

	// fmt.Println(b)
	fd := new(protobuf.FileDescriptorProto)
	if err := proto.Unmarshal(b, fd); err != nil {
		return nil, fmt.Errorf("malformed FileDescriptorProto: %v", err)
	}

	return fd, nil
}
