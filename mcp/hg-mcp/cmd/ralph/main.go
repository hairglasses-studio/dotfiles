package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "ralph is disabled in hg-mcp to conserve Anthropic budget for active coding sessions")
	os.Exit(1)
}
