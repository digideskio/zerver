package routers

import (
	"github.com/cosiner/zerver"

	"testing"
)

func TestInterfaceMatch(t *testing.T) {
	var _ zerver.Router = HostRouter{}
	var _ zerver.RootFilters = HostRootFilters{}
}
