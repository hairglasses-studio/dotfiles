package studio

import (
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestModuleRegistration(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}
