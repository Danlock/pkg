// Package errors prefixes the calling functions name to errors for simpler, smaller traces.
// This package tries to split the difference between github.com/pkg/errors and Go stdlib errors,
// with first class support for log/slog.
package errors

import (
	"errors"
	"fmt"
)

// caller is the number of stack frames to skip when determining the caller's package.func.
const caller = 4

// New creates a new error with the package.func of it's caller prepended.
// It also includes the file and line info of it's caller.
func New(text string) error { return ErrorfWithSkip(caller, text) }

// Errorf is like fmt.Errorf with the "package.func" of it's caller prepended.
// It also includes the file and line info of it's caller.
func Errorf(format string, a ...any) error { return ErrorfWithSkip(caller, format, a...) }

// ErrorfWithSkip is like fmt.Errorf with the "package.func" of the desired caller prepended.
// It also includes the file and line info of it's caller.
func ErrorfWithSkip(skip int, format string, a ...any) error {
	frame := callerFunc(skip)
	merr := metaError{error: fmt.Errorf(prependCaller(format, frame), a...)}
	merr.r.AddAttrs(appendFileToMeta(nil, nil, 0, frame)...)
	return merr
}

// WrapAndPass wraps a typical error func with Wrap and passes the value through unchanged.
func WrapAndPass[T any](val T, err error) (T, error) { return val, WrapfWithSkip(err, caller, "") }

// WrapfAndPass wraps a typical error func with a Wrapf function that passes the value through unchanged.
// WrapMetaCtxAfter contains example usage.
func WrapfAndPass[T any](val T, err error) func(format string, a ...any) (T, error) {
	return func(format string, a ...any) (T, error) {
		return val, WrapfWithSkip(err, caller, format, a...)
	}
}

// Wrap wraps an error with the caller's package.func prepended.
// Similar to github.com/pkg/errors.Wrap and unlike fmt.Errorf it returns nil if err is nil.
// If not wrapping an error from this Go package it also includes the file and line info of it's caller.
func Wrap(err error) error { return WrapfWithSkip(err, caller, "") }

// Wrapf wraps an error with the caller's package.func prepended.
// Similar to github.com/pkg/errors.Wrapf and unlike fmt.Errorf it returns nil if err is nil.
// If not wrapping an error from this Go package it also includes the file and line info of it's caller.
func Wrapf(err error, format string, a ...any) error {
	return WrapfWithSkip(err, caller, format, a...)
}

// WrapfWithSkip wraps an error with the caller's package.func prepended.
// Similar to github.com/pkg/errors.Wrapf and unlike fmt.Errorf it returns nil if err is nil.
// If not wrapping an error from this Go package it also includes the file and line info of it's caller.
// skip is the number of stack frames to skip before recording the function info from runtime.Callers.
func WrapfWithSkip(err error, skip int, format string, a ...any) error {
	if err == nil {
		return nil
	}

	if format == "" {
		format = "%w"
	} else {
		format += " %w"
	}

	frame := callerFunc(skip)
	merr := metaError{error: fmt.Errorf(prependCaller(format, frame), append(a, err)...)}
	merr.r.AddAttrs(appendFileToMeta(nil, err, 0, frame)...)
	return merr
}

// JoinAfter uses error.Join to join deferred errors together.
// JoinAfter set's errPtr to nil if *errPtr and each errFuncs return nil
// errPtr must point to the named error return value from the calling function.
func JoinAfter(errPtr *error, errFuncs ...func() error) {
	if errPtr == nil {
		panic("JoinAfter errPtr must point at the caller function's named return error variable")
	}

	errs := make([]error, 0, len(errFuncs)+1)
	for _, errFunc := range errFuncs {
		if errFunc != nil {
			errs = append(errs, errFunc())
		}
	}

	errs = append(errs, *errPtr)
	// Right now this just returns stdlib errors, but should we wrap it in a metaError if errPtr isn't one?
	*errPtr = errors.Join(errs...)
}

// Into finds the first error in err's chain that matches target type T, and if so, returns it.
// Into is a type-safe alternative to As.
func Into[T any](err error) (val T, ok bool) {
	return val, errors.As(err, &val)
}

// Must is a generic helper, like template.Must, that wraps a call to a function returning (T, error)
// and panics if the error is non-nil.
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

// The following simply call the stdlib so users don't need to include both errors packages.

// ErrUnsupported indicates that a requested operation cannot be performed, because it is unsupported
var ErrUnsupported = errors.ErrUnsupported

// As finds the first error in err's tree that matches target, and if one is found, sets target to that error value and returns true. Otherwise, it returns false.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is reports whether any error in err's tree matches target.
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// Join returns an error that wraps the given errors.
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's type contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
