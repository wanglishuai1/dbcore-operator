package builders

import (
	"fmt"
	"testing"
)

func TestDeployBuilder_Build(t *testing.T) {
	builder, err := NewDeployBuilder("test", "default")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(builder.Replicas(3).Build())
}
