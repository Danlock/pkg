package errors

import (
	"context"
	"log/slog"
	"runtime"
)

type ctxKey struct{}

// WrapMeta is like WrapMetaCtx without the context.
func WrapMeta(err error, meta ...slog.Attr) error {
	if err == nil {
		return nil
	}
	return metaError{error: err, meta: appendFileToMeta(meta, err, caller-1, runtime.Frame{})}
}

func appendMetaFromCtx(ctx context.Context, meta []slog.Attr) []slog.Attr {
	if ctx == nil {
		return meta
	}
	parent, ok := ctx.Value(ctxKey{}).([]slog.Attr)
	if !ok {
		return meta
	}
	return append(meta, parent...)
}

// AddMetaToCtx adds metadata to the context that will be added to the error once WrapMetaCtx is called.
// It creates a new slice each time to prevent data races.
//
// The only way to retrieve the metadata is with UnwrapMeta on an error wrapped with WrapMetaCtx,
// or by just slogging the error which does this internally.
// If you are interested in pulling values out of the context for other purposes,
// take a look at https://github.com/veqryn/slog-context instead.
func AddMetaToCtx(ctx context.Context, meta ...slog.Attr) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, ctxKey{}, appendMetaFromCtx(ctx, meta))
}

// WrapMetaCtx wraps an error with metadata for structured logging.
// Similar to github.com/pkg/errors.Wrap and unlike fmt.Errorf it returns nil if err is nil.
// If not wrapping an error from this Go package it also includes the file and line info of it's caller.
// AddMetaToCtx adds metadata to the ctx which is added to the error, if the context is set.
func WrapMetaCtx(ctx context.Context, err error, meta ...slog.Attr) error {
	if err == nil {
		return nil
	}
	meta = appendFileToMeta(meta, err, caller-1, runtime.Frame{})
	return metaError{error: err, meta: appendMetaFromCtx(ctx, meta)}
}

// WrapMetaCtxAfter is WrapMetaCtx for usage with defer.
// Defer at the top of a function with a named return error variable to wrap any error returned from the function with your desired metadata.
// While safe to call deferred multiple times throughout a function for adding metadata,
// if the context is passed in more than once the context's metadata will be duplicated.
// Depending on your log/slog handler they may get deduplicated, although it will still be a waste of space.
func WrapMetaCtxAfter(ctx context.Context, errPtr *error, meta ...slog.Attr) {
	if errPtr == nil {
		panic("WrapMetaCtxAfter errPtr must point at the caller function's named return error variable")
	}
	if *errPtr == nil {
		return
	}
	meta = appendFileToMeta(meta, *errPtr, caller-1, runtime.Frame{})
	*errPtr = metaError{error: *errPtr, meta: appendMetaFromCtx(ctx, meta)}
}
