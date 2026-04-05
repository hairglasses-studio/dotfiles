# hg-mcp Makefile
# Auto-detect tool paths, preferring local SDK installs

GO := $(shell command -v $(HOME)/go-sdk/go/bin/go 2>/dev/null || command -v go 2>/dev/null)
NODE := $(shell command -v $(HOME)/node-sdk/bin/node 2>/dev/null || command -v node 2>/dev/null)
NPM := $(shell command -v $(HOME)/node-sdk/bin/npm 2>/dev/null || command -v npm 2>/dev/null)

BINARY := hg-mcp-bin
WEB_DIST := internal/web/dist

.PHONY: build build-web test test-short run run-sse run-web clean fmt lint lint-full coverage dry-audit

build: build-web
	$(GO) build -o $(BINARY) ./cmd/hg-mcp

build-web:
	cd web && $(NPM) install && $(NPM) run build
	rm -rf $(WEB_DIST)
	cp -r web/dist $(WEB_DIST)

test:
	$(GO) test ./...

test-short:
	$(GO) test -short ./...

run: build
	./$(BINARY)

run-sse: build
	MCP_MODE=sse PORT=8080 ./$(BINARY)

run-web: build
	MCP_MODE=web PORT=8080 ./$(BINARY)

clean:
	rm -f $(BINARY)
	rm -rf $(WEB_DIST)
	rm -rf web/dist web/node_modules

fmt:
	$(GO) fmt ./...

lint:
	$(GO) vet ./...

lint-full:
	golangci-lint run ./...

coverage:
	$(GO) test -short -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@$(GO) tool cover -func=coverage.out | tail -1

# Show detected tool paths
env:
	@echo "GO:   $(GO)"
	@echo "NODE: $(NODE)"
	@echo "NPM:  $(NPM)"

# DRY violation audit - counts remaining cleanup opportunities
dry-audit:
	@echo "=== DRY Violation Audit ==="
	@echo "GetStringParam calls (total): $$(grep -rn 'GetStringParam' internal/mcp/tools/ --include='*.go' | grep -v _test.go | grep -v helpers.go | wc -l)"
	@echo "RequireStringParam calls (converted): $$(grep -rn 'RequireStringParam' internal/mcp/tools/ --include='*.go' | grep -v _test.go | grep -v helpers.go | wc -l)"
	@echo "json.MarshalIndent in tools: $$(grep -rn 'json.MarshalIndent' internal/mcp/tools/ --include='*.go' | grep -v _test.go | grep -v helpers.go | wc -l)"
	@echo "JSONResult calls: $$(grep -rn 'JSONResult' internal/mcp/tools/ --include='*.go' | grep -v _test.go | grep -v helpers.go | wc -l)"
	@echo "Direct Arguments access: $$(grep -rn 'req\.Params\.Arguments\|request\.Params\.Arguments' internal/mcp/tools/ --include='*.go' | grep -v _test.go | grep -v helpers.go | grep -v registry.go | wc -l)"
	@echo "os.Getenv in tools: $$(grep -rn 'os.Getenv' internal/mcp/tools/ --include='*.go' | grep -v _test.go | wc -l)"

-include $(HOME)/hairglasses-studio/dotfiles/make/pipeline.mk
