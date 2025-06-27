// Package set provides a bog standard set implementation for Go.
package set

// From creates a set from the comparable values passed in.
func From[T comparable](vals ...T) map[T]struct{} {
	m := make(map[T]struct{}, len(vals))
	for _, v := range vals {
		m[v] = struct{}{}
	}
	return m
}
