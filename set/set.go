// Package set provides a bog standard set implementation for Go, just a wrapper around a map[T comparable]struct{}
// Most set methods mutate the set and return it to facilitate method chaining. Not thread safe.
package set

import (
	"iter"
	"maps"
	"slices"
)

// ToSeq concisely converts values into an iter.Seq.
// The values are not made unique.
// Not directly related to sets so may be moved out of this package but
// set.ToSeq(1,2,3) is more concise than slices.Values([]int{1,2,3}).
func ToSeq[T any](v ...T) iter.Seq[T] { return slices.Values(v) }

type Set[T comparable] map[T]struct{}

// From creates a set from the comparable values passed in.
func From[T comparable](vals ...T) Set[T] {
	return make(Set[T], len(vals)).Add(vals...)
}

// FromSeq creates a set from the comparable values passed in via iter.Seq.
func FromSeq[T comparable](vals iter.Seq[T]) Set[T] {
	return make(Set[T]).Union(vals)
}

// All returns an iterator over all elements in the set
func (a Set[T]) All() iter.Seq[T] { return maps.Keys(a) }

// Add adds values to the set
func (a Set[T]) Add(values ...T) Set[T] {
	for _, v := range values {
		a[v] = struct{}{}
	}
	return a
}

// Has returns a boolean indicating whether the set contains all of the values.
func (a Set[T]) Has(values ...T) bool {
	for _, v := range values {
		if _, exists := a[v]; !exists {
			return false
		}
	}
	return true
}

// HasAll returns a boolean indicating whether the set contains all of the sequence.
func (a Set[T]) HasAll(b iter.Seq[T]) bool {
	for v := range b {
		if _, exists := a[v]; !exists {
			return false
		}
	}
	return true
}

// HasAny returns a boolean indicating whether the set contains any of the sequence.
func (a Set[T]) HasAny(b iter.Seq[T]) bool {
	for v := range b {
		if _, exists := a[v]; exists {
			return true
		}
	}
	return false
}

// Union returns the union of the set and sequence
func (a Set[T]) Union(b iter.Seq[T]) Set[T] {
	for v := range b {
		a[v] = struct{}{}
	}
	return a
}

// Difference returns the difference of the set and sequence
func (a Set[T]) Difference(b iter.Seq[T]) Set[T] {
	for v := range b {
		delete(a, v)
	}
	return a
}

// Intersects returns a new set that is the intersection of the set and sequence
func (a Set[T]) Intersects(b iter.Seq[T]) Set[T] {
	in := make(Set[T], len(a))
	for v := range b {
		if _, exists := a[v]; exists {
			in[v] = struct{}{}
		}
	}
	return in
}
