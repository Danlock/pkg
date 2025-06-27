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
	merr := attrError{error: fmt.Errorf(prependCaller(format, frame), a...)}
	merr.r.AddAttrs(appendFileToAttr(nil, nil, 0, frame)...)
	return merr
}

// WrapAndPass wraps a typical error func with Wrap and passes the value through unchanged.
func WrapAndPass[T any](val T, err error) (T, error) { return val, WrapfWithSkip(err, caller, "") }

// WrapfAndPass wraps a typical error func with a Wrapf function that passes the value through unchanged.
// WrapAttrCtxAfter contains example usage.
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
	merr := attrError{error: fmt.Errorf(prependCaller(format, frame), append(a, err)...)}
	merr.r.AddAttrs(appendFileToAttr(nil, err, 0, frame)...)
	return merr
}

// Join returns an error that wraps the given errors.
// Any nil error values are discarded. Join returns nil if every value in errs is nil
func Join(errs ...error) error {
	// wrap the stdlib error so slog.LogValuer can be called
	if err := errors.Join(errs...); err != nil {
		return attrError{error: err}
	}
	return nil
}

// JoinAfter returns an error that wraps the given deferred errors.
// JoinAfter only updates errPtr if one of the errFuncs returned and error.
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

	*errPtr = Join(append(errs, *errPtr)...)
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
