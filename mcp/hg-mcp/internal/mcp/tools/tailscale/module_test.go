package tailscale

import (
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestModuleRegistration(t *testing.T) {
	testutil.AssertModuleInfo(t, testutil.ModuleInfoTest{
		Module:       &Module{},
		ExpectedName: "tailscale",
		MinTools:     3,
		// Mixed CBGroups: CLI tools use "tailscale", API tools use "tailscale-api"
	})
}
