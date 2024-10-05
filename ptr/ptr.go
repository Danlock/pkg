package ptr

// To returns a pointer to value
func To[T any](s T) *T {
	return &s
}

// From dereferences a pointer by returning the zero value if null
func From[T any](p *T) (zero T) {
	if p == nil {
		return zero
	} else {
		return *p
	}
}
