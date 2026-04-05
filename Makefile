.PHONY: test test-lib test-verbose

test: test-lib

test-lib:
	@echo "=== Library tests ==="
	bats tests/lib_*.bats

test-verbose:
	bats --verbose-run --show-output-of-passing-tests tests/*.bats
