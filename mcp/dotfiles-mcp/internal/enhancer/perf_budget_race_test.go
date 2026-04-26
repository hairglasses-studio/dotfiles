//go:build race

package enhancer

import "time"

func largeInputPerformanceBudget() time.Duration {
	return 5 * time.Second
}
