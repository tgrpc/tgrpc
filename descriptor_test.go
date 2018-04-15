package tgrpc

import (
	"testing"
)

func TestGenDescriptorSet(t *testing.T) {
	err := GenDescriptorSet(".", "helloworld.Greeter.pbin", "helloworld/helloworld.proto")
	if err != nil {
		t.Errorf("%s", err)
	}
}
