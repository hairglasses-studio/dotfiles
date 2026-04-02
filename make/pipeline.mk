# pipeline.mk — Standardized build pipeline targets for hairglasses-studio repos.
# Include at the bottom of any repo's Makefile:
#   -include $(HOME)/hairglasses-studio/dotfiles/make/pipeline.mk
#
# Override PIPELINE_GO to use a custom Go wrapper (e.g., ./scripts/dev/go.sh).
# Override PIPELINE_BINARY for repos with non-default binary names.
# Existing repo targets (build, test, vet, lint) are NOT overridden.

PIPELINE_GO      ?= go
PIPELINE_REPO    ?= $(notdir $(CURDIR))
PIPELINE_BINARY  ?= $(PIPELINE_REPO)
PIPELINE_IS_MCP  ?= $(shell test -f .mcp.json && echo yes || echo no)
PIPELINE_MCP_CMD ?= $(shell ls cmd/ 2>/dev/null | grep -q mcp && echo yes || echo no)
PIPELINE_HAS_GO  ?= $(shell test -f go.mod && echo yes || echo no)
PIPELINE_HAS_NODE ?= $(shell test -f package.json && echo yes || echo no)
PIPELINE_HAS_PY  ?= $(shell test -f pyproject.toml && echo yes || echo no)

.PHONY: pipeline-build pipeline-test pipeline-vet pipeline-lint pipeline-check pipeline-info pipeline-detect install-hooks

pipeline-build:
	$(PIPELINE_GO) build ./...

pipeline-test:
	$(PIPELINE_GO) test ./... -count=1 -race

pipeline-vet:
	$(PIPELINE_GO) vet ./...

pipeline-lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || \
	echo "[skip] golangci-lint not installed"

pipeline-check: pipeline-build pipeline-vet pipeline-test

pipeline-detect:
	@printf '{"go":"%s","node":"%s","python":"%s","makefile":"yes"}\n' \
		"$(PIPELINE_HAS_GO)" "$(PIPELINE_HAS_NODE)" "$(PIPELINE_HAS_PY)"

pipeline-info:
	@printf '{"repo":"%s","binary":"%s","is_mcp":"%s","has_mcp_cmd":"%s","go":"%s","node":"%s","python":"%s"}\n' \
		"$(PIPELINE_REPO)" "$(PIPELINE_BINARY)" "$(PIPELINE_IS_MCP)" "$(PIPELINE_MCP_CMD)" \
		"$(PIPELINE_HAS_GO)" "$(PIPELINE_HAS_NODE)" "$(PIPELINE_HAS_PY)"

install-hooks:
	@$(HOME)/hairglasses-studio/dotfiles/scripts/hg-install-hooks.sh
