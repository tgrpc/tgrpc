package tgrpc

import (
	"testing"
)

func TestGenDescriptorSet(t *testing.T) {
	err := GenDescriptorSet("$GOPATH/src/github.com/tgrpc/ngrpc", ".helloworld.Greeter.pbin", "helloworld/helloworld.proto")
	if err != nil {
		t.Errorf("%s", err)
	}
}
