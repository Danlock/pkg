// Personalize the errors stdlib to prepend the calling functions name to errors for simpler, smaller traces
// An example of how this work can be seen at github.com/danlock/gogosseract.
// Example error message from gogosseract:
// gogosseract.NewPool failed worker setup due to gogosseract.(*Pool).runTesseract gogosseract.New gogosseract.Tesseract.createByteView wasm.GetReaderSize io.Reader was empty
package errors

import (
	"errors"
	"fmt"
	"path"
	"runtime"
)

// New creates a new error with the package.func of it's caller prepended.
func New(text string) error {
	return errors.New(prependCaller(text, 3))
}

// Errorf is like fmt.Errorf with the "package.func" of it's caller prepended.
func Errorf(format string, a ...any) error {
	return fmt.Errorf(prependCaller(format, 3), a...)
}

// Errorf is like fmt.Errorf with the "package.func" of the desired caller prepended.
func ErrorfWithSkip(format string, skip int, a ...any) error {
	return fmt.Errorf(prependCaller(format, skip), a...)
}

// Wrap wraps an error with the caller's package.func prepended.
// Similar to github.com/pkg/errors.Wrap it also returns nil if err is nil, unlike fmt.Errorf.
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(prependCaller("%w", 3), err)
}

// Wrapf wraps an error with the caller's package.func prepended.
// Similar to github.com/pkg/errors.Wrap it also returns nil if err is nil, unlike fmt.Errorf.
func Wrapf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	a = append(a, err)
	return fmt.Errorf(prependCaller(format+" %w", 3), a...)
}

func prependCaller(text string, skip int) string {
	var pcs [1]uintptr
	if runtime.Callers(skip, pcs[:]) == 0 {
		return ""
	}
	f := runtime.FuncForPC(pcs[0])
	if f == nil {
		return ""
	}
	// f.Name() gives back something like github.com/danlock/pkg.funcName.
	// with just the package name and the func name, nested errors look more readable by default.
	// We also avoid the ugly giant stack trace cluttering logs and looking similar to panics.
	_, fName := path.Split(f.Name())
	return fmt.Sprint(fName, " ", text)
}

// Into finds the first error in err's chain that matches target type T, and if so, returns it.
//
// Into is a type-safe alternative to As.
func Into[T error](err error) (val T, ok bool) {
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

func Must2[T, U any](val T, val2 U, err error) (T, U) {
	if err != nil {
		panic(err)
	}
	return val, val2
}

// The following simply call the stdlib so users don't need to include both errors packages.

var ErrUnsupported = errors.ErrUnsupported

func As(err error, target any) bool {
	return errors.As(err, target)
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func Join(errs ...error) error {
	return errors.Join(errs...)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}
