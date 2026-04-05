.PHONY: test test-lib test-scripts test-verbose sync clean unmanaged packages

test: test-lib test-scripts

test-lib:
	@echo "=== Library tests ==="
	bats tests/lib_*.bats

test-scripts:
	@echo "=== Script tests ==="
	bats tests/shader_*.bats tests/hg_*.bats

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
