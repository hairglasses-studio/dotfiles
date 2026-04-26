//go:build !race

package enhancer

import "time"

func largeInputPerformanceBudget() time.Duration {
	return time.Second
}
