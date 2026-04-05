package tools

import "github.com/hairglasses-studio/mcpkit/client"

// LazyClient returns a thread-safe lazy-initialized client getter.
// The constructor is called at most once; subsequent calls return the cached result.
//
// Usage in modules (opt-in):
//
//	var getClient = tools.LazyClient(clients.NewTraktorClient)
func LazyClient[T any](constructor func() (T, error)) func() (T, error) {
	return client.LazyClient(constructor)
}
