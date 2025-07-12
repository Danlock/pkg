package errors

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"path"
	"runtime"
	"slices"
	"strings"
)

// AttrError is a structured error using slog.Attr for metadata, similar to log/slog.
// If printed with %+v it will also include the metadata, but by default only the error message is shown.
// The file:line information from the first error in the chain is also included under the DefaultSourceSlogKey.
// Implements slog.LogValuer which logs each slog.Attr in the entire error chain under a slog.GroupValue.
type AttrError interface {
	Attrs() iter.Seq[slog.Attr]
	LogValue() slog.Value
	Unwrap() error
	Error() string
}

var _ = AttrError(attrError{})
var _ = slog.LogValuer(attrError{})
var _ = error(attrError{})

type attrError struct {
	error
	// r is only used to steal log/slog's efficient []slog.Attr implementation
	// that avoids allocations for 5 Attr or less.
	// There is intentionally no way to increase an attrError attrs after it has been created.
	r slog.Record
}

func (e attrError) Unwrap() error  { return e.error }
func (e attrError) String() string { return e.Error() }
func (e attrError) Attrs() iter.Seq[slog.Attr] {
	return func(yield func(slog.Attr) bool) { e.r.Attrs(yield) }
}

// LogValue returns a slog.GroupValue with the file:line information and any existing metadata.
func (e attrError) LogValue() slog.Value {
	metaMap := UnwrapAttr(e)
	meta := make([]slog.Attr, 0, len(metaMap)+1)
	// Order the msg first and the source last for readability.
	if DefaultMsgSlogKey != "" {
		meta = append(meta, slog.String(DefaultMsgSlogKey, e.Error()))
	}
	for k, v := range metaMap {
		if k != DefaultSourceSlogKey {
			meta = append(meta, slog.Attr{Key: k, Value: v})
		}
	}
	// Optionally sort the metadata for tests and anyone else who needs deterministic output.
	if AttrCompareSortFunc != nil {
		slices.SortFunc(meta[1:], AttrCompareSortFunc)
	}
	if DefaultSourceSlogKey != "" {
		meta = append(meta, slog.Attr{Key: DefaultSourceSlogKey, Value: metaMap[DefaultSourceSlogKey]})
	}
	return slog.GroupValue(meta...)
}

// Not sure how I feel about this. I like being able to print all at the metadata in a quick and dirty way
// but if a logger defaults to %+v it would annoyingly duplicate the metadata.
// However as slog is in the stdlib, it's fair to expect other loggers to conform to slog.LogValuer eventually.
func (e attrError) Format(s fmt.State, verb rune) {
	if verb == 'v' && s.Flag('+') {
		// This outputs all metadata for ease of debugging but really... just use log/slog.
		_, _ = io.WriteString(s, e.LogValue().String())
	} else {
		_, _ = io.WriteString(s, e.Error())
	}
}

func callerFunc(skip int) runtime.Frame {
	var pcs [1]uintptr
	if runtime.Callers(skip, pcs[:]) == 0 {
		return runtime.Frame{}
	}
	frames := runtime.CallersFrames(pcs[:])
	if frames == nil {
		return runtime.Frame{}
	}
	frame, _ := frames.Next()
	if DefaultPackagePrefix == "" {
		return frame
	}
	// Trim the file path down to just what we need to identify it from the package name.
	_, after, _ := strings.Cut(frame.File, DefaultPackagePrefix)
	if len(after) > 0 {
		frame.File = DefaultPackagePrefix + after
	}
	return frame
}

func prependCaller(text string, f runtime.Frame) string {
	if f.Function == "" {
		return text
	}
	// runtime.Frame.Function gives back something like github.com/danlock/pkg.funcName.
	// with just the package name and the func name, nested errors look more readable by default.
	// We also avoid the ugly giant stack trace cluttering logs and looking similar to panics.
	// Now that the file:line of the original error is also within the metadata,
	// trimming the fat makes errors easier to parse at a glance.
	_, fName := path.Split(f.Function)
	return fmt.Sprint(fName, " ", text)
}

// appendFileToAttr appends the file and line info of the caller to the metadata if it's the first error from this package in the chain.
// If skip is greater than 0 it reads the frame instead of using the passed in frame.
func appendFileToAttr(meta []slog.Attr, err error, skip int, frame runtime.Frame) []slog.Attr {
	if DefaultSourceSlogKey == "" {
		return meta
	}
	if _, exist := Into[attrError](err); exist {
		return meta
	}
	if skip > 0 {
		frame = callerFunc(skip)
	}
	return append(meta, slog.String(DefaultSourceSlogKey, fmt.Sprintf("%s:%d", frame.File, frame.Line)))
}

// updateAttrMapFromErr adds err's metadata into the given map.
// This deduplicates metadata across the error chain, which allows multiple deferred WrapAttrCtxAfter calls
// in a single function for example, which would usually duplicate the fields added to the context.
func updateAttrMapFromErr(err error, meta map[string]slog.Value) {
	// errors.As only returns the first error in an errors.Join error, so we handle those recursively beforehand
	jerr, ok := Into[interface{ Unwrap() []error }](err)
	if ok {
		for _, e := range jerr.Unwrap() {
			updateAttrMapFromErr(e, meta)
		}
	}
	// errors.As will also end up grabbing one of the joined errors, so we output to a map to avoid duplication.
	var merr AttrError
	for errors.As(err, &merr) {
		for attr := range merr.Attrs() {
			meta[attr.Key] = attr.Value
		}
		err = errors.Unwrap(merr)
	}
}

// UnwrapAttr returns a map around an error chain's metadata.
// If the error lacks metadata an empty map is returned.
// Since errors in this package implement slog.LogValuer you don't need to use this to pass slog.Attr to slog.Log.
//
// Structured errors can be introspected and handled differently as needed.
func UnwrapAttr(err error) map[string]slog.Value {
	meta := make(map[string]slog.Value)
	updateAttrMapFromErr(err, meta)
	return meta
}

// Get retrieves a value from the error's metadata.
// Returns an error if the key doesn't exist or the type is incorrect instead of panicking like the slog.Value methods do.
// This can be used for retrieving an error code added to the error chain, for example Get[uint64](meta, "http.code")
//
// Note that meta["http.code"].Uint64() is more efficient in regards to allocations.
func Get[T any](meta map[string]slog.Value, key string) (val T, err error) {
	// Was tempted to return a structured error with key as an attr,
	// but I feel libraries should stick with stdlib errors, fmt.Errorf and if needed, a custom error type.
	// Structured error attributes are only an application concern, especially since duplicate keys aren't supported.
	// What if the application wants to use the same key for a different value?
	// The file:line info is also most useful when it points to where an application calls a library, instead of deep within the library itself.
	//
	// If regardless you want to use this structured error library, at least try to minimize duplicate key issues by grouping under your package name:
	// defer WrapAttrCtxAfter(nil, &err, slog.Group("github.com/your/package", "key", val))
	if meta == nil {
		return val, errors.New("errors.Get called with nil meta")
	}
	sVal, ok := meta[key]
	if !ok {
		return val, fmt.Errorf("errors.Get key %s not found", key)
	}
	val, ok = sVal.Any().(T)
	if !ok {
		return val, fmt.Errorf("errors.Get key %s had kind %s and not type %T", key, sVal.Kind(), val)
	}
	return val, nil
}
