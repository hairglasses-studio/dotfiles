.PHONY: build test vet contract-snapshot contract-check contract-diff host-smoke host-smoke-strict release-parity publish-check

GO ?= go
GO_RUN := env GOWORK=off $(GO)

build:
	$(GO_RUN) build -o dotfiles-mcp .

test:
	$(GO_RUN) test ./... -count=1

vet:
	$(GO_RUN) vet ./...

contract-snapshot:
	$(GO_RUN) run . --contract-write

contract-check:
	$(GO_RUN) run . --contract-check

contract-diff:
	bash ./scripts/contract-diff-summary.sh

host-smoke:
	bash ./scripts/host-smoke.sh

host-smoke-strict:
	bash ./scripts/host-smoke.sh --strict-missing --strict-skip

release-parity:
	bash ./scripts/release-parity.sh

publish-check: vet test contract-check release-parity

-include $(HOME)/hairglasses-studio/dotfiles/make/pipeline.mk
