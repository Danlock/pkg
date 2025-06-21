// Package ptr provides utility functions for working with pointers.
package ptr

import (
	"reflect"
)

// To returns a pointer to value
func To[T any](s T) *T {
	return &s
}

// From dereferences a pointer by returning the zero value if null
func From[T any](p *T) (zero T) {
	if p == nil {
		return zero
	}
	return *p
}

// IsInterfaceNil checks if either an interface or it's underlying concrete value is nil.
// If the type can't be nil, it return's false.
func IsInterfaceNil(i any) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	default:
		return false
	}
}

// Or returns the first of its arguments that is not equal to the zero value.
// If no argument is non-zero, it returns the zero value.
// Essentially cmp.Or that rejects nil interfaces.
func Or[T comparable](vals ...T) T {
	var zero T
	for _, val := range vals {
		if val != zero && !IsInterfaceNil(val) {
			return val
		}
	}
	return zero
}
