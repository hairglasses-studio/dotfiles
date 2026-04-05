package prolink

import (
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleInfo(t, testutil.ModuleInfoTest{
		Module:       &Module{},
		ExpectedName: "prolink",
		MinTools:     11,
	})
}
