.PHONY: test test-lib test-scripts test-verbose

test: test-lib test-scripts

test-lib:
	@echo "=== Library tests ==="
	bats tests/lib_*.bats

test-scripts:
	@echo "=== Script tests ==="
	bats tests/shader_*.bats tests/hg_*.bats

test-verbose:
	bats --verbose-run --show-output-of-passing-tests tests/*.bats
