BASH_SCRIPTS := $(shell awk 'FNR == 1 && $$0 ~ /bash/ { print FILENAME }' install.sh scripts/*.sh scripts/lib/*.sh)

.PHONY: test test-lib test-scripts test-verbose sync clean unmanaged packages lint check-shaders validate-shaders

test: test-lib test-scripts

lint:
	@echo "=== Shell Syntax Checks ==="
	bash -n $(BASH_SCRIPTS)
	@echo "=== Shellcheck ==="
	shellcheck $(BASH_SCRIPTS) || true
	@echo "=== JSON Validation ==="
	find . -maxdepth 2 -name "*.json" -not -path "*/node_modules/*" -exec jq . {} + >/dev/null
	@echo "=== Shader Consistency ==="
	scripts/check-shader-consistency.sh

validate-shaders:
	@echo "=== DarkWindow Shader Validation ==="
	scripts/validate-darkwindow-shaders.sh --baseline 120

check-shaders:
	scripts/check-shader-consistency.sh

test-lib:
	@echo "=== Library tests ==="
	bats tests/lib_*.bats

test-scripts:
	@echo "=== Script tests ==="
	bats tests/shader_*.bats tests/hg_*.bats tests/install_check_worktree.bats tests/repo_smoke.bats tests/theme_sync.bats

test-verbose:
	bats --verbose-run --show-output-of-passing-tests tests/*.bats

# Package management via metapac
sync:
	metapac sync

clean:
	metapac clean

unmanaged:
	metapac unmanaged

packages: sync
