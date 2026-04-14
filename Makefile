.PHONY: test test-lib test-scripts test-verbose sync clean unmanaged packages lint

test: test-lib test-scripts

lint:
	@echo "=== Shell Syntax Checks ==="
	bash -n scripts/*.sh scripts/lib/*.sh
	@echo "=== Shellcheck ==="
	shellcheck scripts/*.sh scripts/lib/*.sh || true
	@echo "=== JSON Validation ==="
	find . -maxdepth 2 -name "*.json" -not -path "*/node_modules/*" -exec jq . {} + >/dev/null

test-lib:
	@echo "=== Library tests ==="
	bats tests/lib_*.bats

test-scripts:
	@echo "=== Script tests ==="
	bats tests/shader_*.bats tests/hg_*.bats tests/install_check_worktree.bats tests/repo_smoke.bats

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
