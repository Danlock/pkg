package errors

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
)

// DefaultFileSlogKey is the default slog.Attr key used for file:line information when an error is printed via log/slog.
// If set to "", file:line metadata will not be included in errors.
var DefaultFileSlogKey = "file"

// DefaultFilePackagePrefix trims the file:line path of the build location using this package name.
// With Go modules it's updated automatically, but without Go modules it defaults to github.com/ and may need to be updated for your project.
// If set to "" the file metadata is not trimmed at all.
//
// As an example by default we trim /home/dan/go/src/github.com/danlock/pkg/errors/meta_test.go:30 down to github.com/danlock/pkg/errors/meta_test.go:30.
var DefaultFilePackagePrefix = "github.com/"

// ShouldSortUnwrapMeta controls whether UnwrapMeta, and therefore the slog GroupValue log output, will be sorted by key for determinism.
var ShouldSortUnwrapMeta = false

func init() {
	// Automatically configure DefaultFilePackagePrefix with Go modules.
	bi, ok := debug.ReadBuildInfo()
	if !ok || bi == nil || len(bi.Path) == 0 {
		return
	}
	before, _, ok := strings.Cut(bi.Path, "/")
	if ok {
		DefaultFilePackagePrefix = before + "/"
	}
}

// metaError is a structured stdlib Go error using slog.Attr for metadata.
// If printed with %+v it will also include the metadata, but by default only the error message is shown.
// It will also include the file:line information from the first error in the chain under the DefaultFileSlogKey.
// Meant for use with log/slog where everything converts to a slog.GroupValue when logged.
type metaError struct {
	error
	meta []slog.Attr
}

// UnwrapMeta calls UnwrapMeta on itself, for external packages that need to access this error chain's metadata without relying on this package directly.
func (e metaError) UnwrapMeta() []slog.Attr { return UnwrapMeta(e) }
func (e metaError) Unwrap() error           { return e.error }
func (e metaError) String() string          { return e.Error() }

// LogValue logs the error with the file:line information and any existing metadata.
func (e metaError) LogValue() slog.Value {
	return slog.GroupValue(append(
		UnwrapMeta(e), slog.String("msg", e.Error()))...)
}

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
// However as slog is in the stdlib, it's fair to expect other loggers to conform to slog.LogValuer eventually.
func (e metaError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			// This outputs all metadata for ease of debugging but really... just use log/slog.
			_, _ = io.WriteString(s, e.Error()+" "+stringifyAttr(UnwrapMeta(e)))
			return
		}
		fallthrough
	case 's':
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
	if DefaultFilePackagePrefix == "" {
		return frame
	}
	// Trim the file path down to just what we need to identify it from the package name.
	_, after, _ := strings.Cut(frame.File, DefaultFilePackagePrefix)
	if len(after) > 0 {
		frame.File = DefaultFilePackagePrefix + after
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

// appendFileToMeta appends the file and line info of the caller to the metadata if it's the first error from this package in the chain.
// If skip is greater than 0 it reads the frame instead of using the passed in frame.
func appendFileToMeta(meta []slog.Attr, err error, skip int, frame runtime.Frame) []slog.Attr {
	if DefaultFileSlogKey == "" {
		return meta
	}
	if _, exist := Into[metaError](err); exist {
		return meta
	}
	if skip > 0 {
		frame = callerFunc(skip)
	}
	return append(meta, slog.String(DefaultFileSlogKey, fmt.Sprintf("%s:%d", frame.File, frame.Line)))
}

type joinedError interface {
	Unwrap() []error
	Error() string
}

// updateMetaMapFromErr adds err's metadata into the given map.
// This deduplicates metadata across the error chain, which allows multiple deferred WrapMetaCtxAfter calls
// in a single function for example, which would usually duplicate the fields added to the context.
func updateMetaMapFromErr(err error, meta map[string]slog.Value) {
	// errors.As only returns the first error in an errors.Join error, so we handle those recursively beforehand
	if jerr, ok := Into[joinedError](err); ok {
		for _, e := range jerr.Unwrap() {
			updateMetaMapFromErr(e, meta)
		}
	}
	// errors.As will also end up grabbing one of the joined errors, so we output to a map to avoid duplication.
	var merr metaError
	for errors.As(err, &merr) {
		for _, attr := range merr.meta {
			meta[attr.Key] = attr.Value
		}
		err = errors.Unwrap(merr)
	}
}

// UnwrapMeta pulls metadata from every error in the chain for structured logging purposes.
// Errors in this package implement slog.LogValuer and automatically include the metadata when used with slog.Log.
// This function is mainly exposed for use with loggers other than log/slog.
// Duplicate keys across the error chain are not supported.
// As this function internally uses a map, the returned slice order is non deterministic unless ShouldSortUnwrapMeta true.
func UnwrapMeta(err error) []slog.Attr {
	metaMap := make(map[string]slog.Value)
	updateMetaMapFromErr(err, metaMap)
	meta := make([]slog.Attr, 0, len(metaMap))
	for k, v := range metaMap {
		meta = append(meta, slog.Attr{Key: k, Value: v})
	}
	if ShouldSortUnwrapMeta {
		slices.SortFunc(meta, func(a, b slog.Attr) int { return cmp.Compare(a.Key, b.Key) })
	}
	return meta
}

// UnwrapMetaMap returns a map around an error's metadata.
// If the error lacks metadata an empty map is returned.
//
// Structured errors can be introspected and handled differently as needed.
// Duplicate keys across the error chain are not allowed.
//
// Seriously consider a sentinel error or custom error type before reaching for this.
// For example open source libraries would be better off publicly exposing custom error types for type safety.
//
// Using const keys is strongly recommended to avoid typos.
func UnwrapMetaMap(err error) map[string]slog.Value {
	meta := make(map[string]slog.Value)
	updateMetaMapFromErr(err, meta)
	return meta
}
