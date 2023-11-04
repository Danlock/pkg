// Personalize the errors stdlib to prepend the calling functions name to errors for simple traces
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

func New(text string) error {
	return errors.New(prependCaller(text, 2) + text)
}

func Errorf(format string, a ...any) error {
	return fmt.Errorf(prependCaller(format, 2), a...)
}

func ErrorfWithSkip(format string, skip int, a ...any) error {
	return fmt.Errorf(prependCaller(format, skip), a...)
}

func prependCaller(text string, skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	f := runtime.FuncForPC(pc)
	if f == nil {
		return ""
	}
	// f.Name() gives back something like github.com/danlock/pkg.funcName.
	// with just the package name and the func name, nested errors look more readable by default.
	// We also avoid an ugly giant stack trace that won't always get printed out.
	_, fName := path.Split(f.Name())
	return fmt.Sprint(fName, " ", text)
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
