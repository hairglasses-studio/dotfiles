package gtasks

import (
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleInfo(t, testutil.ModuleInfoTest{
		Module:       &Module{},
		ExpectedName: "gtasks",
		MinTools:     1,
		CBGroup:      "google",
	})
}
