package inventory

import (
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}
