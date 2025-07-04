package errors

import (
	"errors"
	"log/slog"
	"runtime/debug"
	"strings"
)

var (
	// DefaultSourceSlogKey is the default slog.Attr key used for file:line information when an error is printed.
	// If set to "", file:line metadata will not be included in errors.
	DefaultSourceSlogKey = slog.SourceKey

	// DefaultMsgSlogKey is the default slog.Attr key used for the error message when an error is printed.
	// If set to "", the error message will not be included in the slog.LogValuer group.
	DefaultMsgSlogKey = slog.MessageKey

	// DefaultPackagePrefix controls the trimming of the build location out of the file:line source.
	// With Go modules it's updated automatically, but without Go modules it defaults to github.com/ and may need to be updated for your project.
	// If set to "" the source path is not trimmed at all.
	//
	// trimming example: /home/dan/go/src/github.com/danlock/pkg/errors/attr_test.go:30 -> github.com/danlock/pkg/errors/attr_test.go:30
	DefaultPackagePrefix = "github.com/"

	// ShouldSortAttr controls whether LogValue output will be sorted by key for determinism.
	// Regardless of this value msg will be first and source will be last.
	ShouldSortAttr = false
)

func init() {
	// Use Go modules to set DefaultPackagePrefix.
	bi, ok := debug.ReadBuildInfo()
	if !ok || bi == nil || len(bi.Path) == 0 {
		return
	}
	before, _, ok := strings.Cut(bi.Path, "/")
	if ok {
		DefaultPackagePrefix = before + "/"
	}
}

// The following simply call the stdlib so users don't need to include both errors packages.

// ErrUnsupported indicates that a requested operation cannot be performed, because it is unsupported. Calls stdlib errors.ErrUnsupported
var ErrUnsupported = errors.ErrUnsupported

// As finds the first error in err's tree that matches target, and if one is found, sets target to that error value and returns true. Otherwise, it returns false.
// Calls stdlib errors.As
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is reports whether any error in err's tree matches target.
// Calls stdlib errors.Is
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's type contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
// Calls stdlib errors.Unwrap
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
