package main

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/mcpkit/prompts"
	"github.com/hairglasses-studio/mcpkit/registry"
	"github.com/hairglasses-studio/mcpkit/resources"
)

func TestBuildDotfilesPromptRegistry(t *testing.T) {
	promptReg := buildDotfilesPromptRegistry()
	if promptReg.PromptCount() != 8 {
		t.Fatalf("expected 8 prompts, got %d", promptReg.PromptCount())
	}
	if _, ok := promptReg.GetPrompt("dotfiles_audit_fleet"); !ok {
		t.Fatal("expected dotfiles_audit_fleet prompt to be registered")
	}
	if _, ok := promptReg.GetPrompt("dotfiles_repair_config"); !ok {
		t.Fatal("expected dotfiles_repair_config prompt to be registered")
	}
	if _, ok := promptReg.GetPrompt("dotfiles_cleanup_repo_hygiene"); !ok {
		t.Fatal("expected dotfiles_cleanup_repo_hygiene prompt to be registered")
	}
}

func TestBuildDotfilesResourceRegistry(t *testing.T) {
	reg := registry.NewToolRegistry()
	promptReg := buildDotfilesPromptRegistry()
	resReg := buildDotfilesResourceRegistry(reg, promptReg)
	if resReg.ResourceCount() != 16 {
		t.Fatalf("expected 16 resources, got %d", resReg.ResourceCount())
	}
	if _, ok := resReg.GetResource("dotfiles://server/overview"); !ok {
		t.Fatal("expected dotfiles overview resource to be registered")
	}
	if _, ok := resReg.GetResource("dotfiles://catalog/workflows"); !ok {
		t.Fatal("expected dotfiles workflow catalog resource to be registered")
	}
}

func TestDotfilesOverviewResource(t *testing.T) {
	reg := registry.NewToolRegistry()
	promptReg := buildDotfilesPromptRegistry()
	resModule := &dotfilesResourceModule{reg: reg, promptReg: promptReg}

	var overview resources.ResourceDefinition
	for _, rd := range resModule.Resources() {
		if rd.Resource.URI == "dotfiles://server/overview" {
			overview = rd
			break
		}
	}

	out, err := overview.Handler(context.Background(), mcp.ReadResourceRequest{})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 resource contents entry, got %d", len(out))
	}
	text, ok := out[0].(mcp.TextResourceContents)
	if !ok {
		t.Fatalf("expected text resource contents, got %T", out[0])
	}
	if text.Text == "" {
		t.Fatal("expected non-empty overview text")
	}
	if !containsText(text.Text, "dotfiles://catalog/workflows") {
		t.Fatal("expected overview to mention the workflow catalog")
	}
}

func TestDotfilesPromptHandler(t *testing.T) {
	promptModule := &dotfilesPromptModule{}

	var pd prompts.PromptDefinition
	for _, candidate := range promptModule.Prompts() {
		if candidate.Prompt.Name == "dotfiles_triage_desktop" {
			pd = candidate
			break
		}
	}

	result, err := pd.Handler(context.Background(), mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Arguments: map[string]string{"symptom": "eww bar missing"},
		},
	})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result == nil || len(result.Messages) == 0 {
		t.Fatal("expected prompt result with messages")
	}
}

func TestDotfilesWorkflowCatalogResource(t *testing.T) {
	reg := registry.NewToolRegistry()
	promptReg := buildDotfilesPromptRegistry()
	resModule := &dotfilesResourceModule{reg: reg, promptReg: promptReg}

	var catalog resources.ResourceDefinition
	for _, rd := range resModule.Resources() {
		if rd.Resource.URI == "dotfiles://catalog/workflows" {
			catalog = rd
			break
		}
	}

	out, err := catalog.Handler(context.Background(), mcp.ReadResourceRequest{})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	text := out[0].(mcp.TextResourceContents)
	if !containsText(text.Text, `"config_repair"`) {
		t.Fatalf("expected config_repair in workflow catalog: %s", text.Text)
	}
	if !containsText(text.Text, `"repo_hygiene"`) {
		t.Fatalf("expected repo_hygiene in workflow catalog: %s", text.Text)
	}
	if !containsText(text.Text, `"dotfiles_repair_config"`) {
		t.Fatalf("expected dotfiles_repair_config in workflow catalog: %s", text.Text)
	}
	if !containsText(text.Text, `"dotfiles_cleanup_repo_hygiene"`) {
		t.Fatalf("expected dotfiles_cleanup_repo_hygiene in workflow catalog: %s", text.Text)
	}
}

func TestDotfilesPrioritiesResource(t *testing.T) {
	reg := registry.NewToolRegistry()
	promptReg := buildDotfilesPromptRegistry()
	resModule := &dotfilesResourceModule{reg: reg, promptReg: promptReg}

	var priorities resources.ResourceDefinition
	for _, rd := range resModule.Resources() {
		if rd.Resource.URI == "dotfiles://catalog/priorities" {
			priorities = rd
			break
		}
	}

	out, err := priorities.Handler(context.Background(), mcp.ReadResourceRequest{})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	text := out[0].(mcp.TextResourceContents)
	if !containsText(text.Text, `"missing_front_door_count": 0`) {
		t.Fatalf("expected zero missing front doors: %s", text.Text)
	}
	if !containsText(text.Text, `"workflow_count": 8`) {
		t.Fatalf("expected workflow count in priorities resource: %s", text.Text)
	}
}

func containsText(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
