// Package set provides a bog standard set implementation for Go, just a wrapper around a map[T comparable]struct{}
// Most set methods mutate the set and return it to facilitate method chaining. Not thread safe.
package set

import (
	"iter"
	"slices"
	"testing"

	"github.com/danlock/pkg/test"
)

func Test(t *testing.T) {
	names := []string{"Anton", "Brock", "Chairo"}

	nameSet := From(names...)
	nameSet = FromSeq(slices.Values(names))

	intIter := func(i iter.Seq[int]) {}
	intIter(ToSeq(1, 2, 3))

	combined := nameSet.
		Union(ToSeq("Dave", "Eve", "Anton")).
		Add("Joe", "Frank").
		Difference(ToSeq("Alice", "Eve", "whodat")).
		Intersects(ToSeq("Brock", "Dave"))

	test.Truth(t, !combined.Has("Alice"), "Expected Alice to not be in the set")
	test.Truth(t, combined.Has("Dave"), "Expected Dave to be in the set")
	test.Truth(t, combined.HasAny(ToSeq("Alice", "Brock")), "Expected Brock to be in the set")
	test.Truth(t, !combined.HasAll(ToSeq("Alice", "Dave")), "Expected Alice to not be in the set")
}
