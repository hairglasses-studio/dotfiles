package testutil

import (
	"context"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// stubModule implements tools.ToolModule for testing the helpers themselves.
type stubModule struct {
	name  string
	desc  string
	tools []tools.ToolDefinition
}

func (m *stubModule) Name() string                    { return m.name }
func (m *stubModule) Description() string             { return m.desc }
func (m *stubModule) Tools() []tools.ToolDefinition   { return m.tools }

var noopHandler = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{}, nil
}

func TestAssertModuleValid_Success(t *testing.T) {
	mod := &stubModule{
		name: "test",
		desc: "A test module",
		tools: []tools.ToolDefinition{
			{
				Tool: mcp.Tool{
					Name:        "test_action",
					Description: "Does something",
				},
				Handler:             noopHandler,
				CircuitBreakerGroup: "test",
			},
		},
	}

	// Should not panic or fail
	AssertModuleValid(t, mod)
}

func TestAssertModuleValid_MultipleTools(t *testing.T) {
	mod := &stubModule{
		name: "multi",
		desc: "Multi-tool module",
		tools: []tools.ToolDefinition{
			{
				Tool: mcp.Tool{
					Name:        "multi_list",
					Description: "List items",
				},
				Handler:             noopHandler,
				CircuitBreakerGroup: "multi",
			},
			{
				Tool: mcp.Tool{
					Name:        "multi_get",
					Description: "Get item",
				},
				Handler:             noopHandler,
				CircuitBreakerGroup: "multi",
			},
		},
	}

	AssertModuleValid(t, mod)
}

func TestAssertModuleInfo_Success(t *testing.T) {
	mod := &stubModule{
		name: "mymod",
		desc: "My module",
		tools: []tools.ToolDefinition{
			{
				Tool: mcp.Tool{
					Name:        "mymod_action",
					Description: "An action",
				},
				Handler:             noopHandler,
				CircuitBreakerGroup: "mymod",
			},
			{
				Tool: mcp.Tool{
					Name:        "mymod_status",
					Description: "Status check",
				},
				Handler:             noopHandler,
				CircuitBreakerGroup: "mymod",
			},
		},
	}

	AssertModuleInfo(t, ModuleInfoTest{
		Module:       mod,
		ExpectedName: "mymod",
		MinTools:     1,
		CBGroup:      "mymod",
	})
}

func TestAssertModuleInfo_NoCBGroup(t *testing.T) {
	mod := &stubModule{
		name: "nocb",
		desc: "No CB module",
		tools: []tools.ToolDefinition{
			{
				Tool: mcp.Tool{
					Name:        "nocb_action",
					Description: "An action",
				},
				Handler:             noopHandler,
				CircuitBreakerGroup: "anything",
			},
		},
	}

	// When CBGroup is empty in the config, it should not check CB groups
	AssertModuleInfo(t, ModuleInfoTest{
		Module:       mod,
		ExpectedName: "nocb",
		MinTools:     1,
		CBGroup:      "", // empty = skip CB check
	})
}

func TestAssertModuleInfo_MinToolsMet(t *testing.T) {
	mod := &stubModule{
		name: "minmod",
		desc: "Min tools module",
		tools: []tools.ToolDefinition{
			{
				Tool:                mcp.Tool{Name: "t1", Description: "d1"},
				Handler:             noopHandler,
				CircuitBreakerGroup: "minmod",
			},
			{
				Tool:                mcp.Tool{Name: "t2", Description: "d2"},
				Handler:             noopHandler,
				CircuitBreakerGroup: "minmod",
			},
			{
				Tool:                mcp.Tool{Name: "t3", Description: "d3"},
				Handler:             noopHandler,
				CircuitBreakerGroup: "minmod",
			},
		},
	}

	// MinTools = 3 should pass with exactly 3 tools
	AssertModuleInfo(t, ModuleInfoTest{
		Module:       mod,
		ExpectedName: "minmod",
		MinTools:     3,
		CBGroup:      "minmod",
	})
}
