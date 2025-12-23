.PHONY: help diagrams bootstrap validate

help:
	@echo "Targets:"
	@echo "  diagrams  - Render Mermaid diagrams to diagrams/_rendered (Docker/Podman/mmdc)"
	@echo "  bootstrap - Enable hooks, patch README badges, render diagrams"
	@echo "  validate  - Validate DMX patch (tools/validate_patch.py)"

diagrams:
	./scripts/render_diagrams.sh

bootstrap:
	./scripts/bootstrap.sh

validate:
	python3 tools/validate_patch.py

.PHONY: dev-check
dev-check:
	./scripts/dev-check.sh
