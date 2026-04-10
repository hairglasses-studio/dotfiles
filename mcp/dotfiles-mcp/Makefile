.PHONY: build test vet

GO ?= go
GO_RUN := env GOWORK=off $(GO)

build:
	$(GO_RUN) build -o dotfiles-mcp .

test:
	$(GO_RUN) test ./... -count=1

vet:
	$(GO_RUN) vet ./...

-include $(HOME)/hairglasses-studio/dotfiles/make/pipeline.mk
