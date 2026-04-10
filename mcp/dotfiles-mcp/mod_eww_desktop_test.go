package main

import (
	"testing"

	"github.com/hairglasses-studio/mcpkit/mcptest"
	"github.com/hairglasses-studio/mcpkit/registry"
)

func TestEwwDesktopModuleRegistration(t *testing.T) {
	m := &EwwDesktopModule{}
	tools := m.Tools()
	if len(tools) != 2 {
		t.Fatalf("expected 2 eww desktop tools, got %d", len(tools))
	}

	reg := registry.NewToolRegistry()
	reg.RegisterModule(m)
	srv := mcptest.NewServer(t, reg)

	for _, want := range []string{
		"dotfiles_eww_inspect",
		"dotfiles_eww_reload",
	} {
		if !srv.HasTool(want) {
			t.Errorf("missing tool: %s", want)
		}
	}
}
