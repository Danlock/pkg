package errors

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// metaErr is a regular error with metadata stored in slog.Attr for structured logging.
// If printed with %+v it will also include the metadata, but by default only the error message is shown.
// Use UnwrapMeta() -> slog.LogAttr() to read all metadata attached to a chain of errors.
type metaErr struct {
	error
	meta []slog.Attr
}

func (e metaErr) Unwrap() error  { return e.error }
func (e metaErr) String() string { return e.Error() }

// LogValue overwrites slog's default logging behaviour of %+v which would duplicate the metadata.
func (e metaErr) LogValue() slog.Value { return slog.StringValue(e.Error()) }

func stringifyAttr(meta []slog.Attr) string {
	if len(meta) == 0 {
		return ""
	}

	var all strings.Builder
	all.WriteString("{")
	for i, attr := range meta {
		all.WriteString(attr.String())
		if i < len(meta)-1 {
			all.WriteString(",")
		}
	}
	all.WriteString("}")
	return all.String()

}

// Not sure how I feel about this. I like being able to print all at the metadata in a quick and dirty way
// but if a logger defaults to %+v it would annoyingly duplicate the metadata.
func (e metaErr) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			// This outputs all metadata for ease of debugging but UnwrapMeta() -> slog.LogAttr() is preferred for structured logging.
			io.WriteString(s, e.Error()+" "+stringifyAttr(UnwrapMetaSansErr(e)))
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.Error())
	}
}

// UnwrapMeta pulls metadata from every error in the chain for structured logging purposes.
// The error itself is also included as slog.Any("err", err) for ease of use with slog.LogAttrs.
func UnwrapMeta(err error) []slog.Attr {
	return append(UnwrapMetaSansErr(err), slog.Any("err", err))

}

// UnwrapMetaSansErr pulls metadata from every error in the chain for structured logging purposes.
// It doesn't include the error itself, in case you want a different error field or something.
func UnwrapMetaSansErr(err error) (meta []slog.Attr) {
	var se metaErr
	for errors.As(err, &se) {
		meta = append(meta, se.meta...)
		err = errors.Unwrap(se)
	}
	return meta
}

// WrapMeta wraps an error with metadata for structured logging, if it exists.
func WrapMeta(err error, meta ...slog.Attr) error {
	if err == nil {
		return nil
	}

	return metaErr{error: err, meta: meta}
}
