package host

import (
	"github.com/cosiner/zerver"

	"testing"
)

func TestInterfaceMatch(t *testing.T) {
	var _ zerver.Router = NewRouter()
	var _ zerver.RootFilters = NewRootFilters()
}
