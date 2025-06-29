package errors

import (
	"context"
	"log/slog"
	"runtime"
)

type ctxKey struct{}

// WrapAttr is WrapAttrCtx without the context.
func WrapAttr(err error, meta ...slog.Attr) error {
	if err == nil {
		return nil
	}
	meta = appendFileToAttr(meta, err, caller-1, runtime.Frame{})
	return maybeWrapAttrError(err, meta)
}

func appendAttrFromCtx(ctx context.Context, meta []slog.Attr) []slog.Attr {
	if ctx == nil {
		return meta
	}
	parent, ok := ctx.Value(ctxKey{}).([]slog.Attr)
	if !ok {
		return meta
	}
	return append(meta, parent...)
}

// maybeWrapAttrError wraps the error only if we have metadata or the error isn't a AttrError.
func maybeWrapAttrError(err error, meta []slog.Attr) error {
	// Wrapping the error with a attrError without metadata or a better error message
	// just bloats the linked list of errors without any benefit.
	//
	// However we do need a AttrError at the top of the error chain so our slog.LogValuer will be called.
	_, isAttr := err.(AttrError)
	if len(meta) == 0 && isAttr {
		return err
	}
	merr := attrError{error: err}
	merr.r.AddAttrs(meta...)
	return merr
}

// AddAttrToCtx adds metadata to the context that will be added to the error once WrapAttrCtx is called.
// It appends to any existing metadata in the context.
//
// The only way to retrieve the metadata is with UnwrapAttr on an error wrapped with WrapAttrCtx,
// or by just slogging the error which handles this internally.
//
// If you are interested in pulling values out of the context for other purposes,
// take a look at https://github.com/veqryn/slog-context instead.
func AddAttrToCtx(ctx context.Context, meta ...slog.Attr) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, ctxKey{}, appendAttrFromCtx(ctx, meta))
}

// WrapAttrCtx wraps an error with metadata for structured logging.
// Similar to github.com/pkg/errors.Wrap and unlike fmt.Errorf it returns nil if err is nil.
//
// If not wrapping an error from this Go package it also includes the file and line info of it's caller.
// AddAttrToCtx adds metadata to the ctx which will be added to the error, if the context is set.
//
// If 0 metadata will be included with the error, i.e both context is nil and meta is empty,
// the original error will be returned to avoid bloating the error chain.
//
// Note that the slog output contains 2 keys by default, DefaultSourceSlogKey and DefaultMsgSlogKey,
// which use slog's standard "source" and "msg". Duplicate keys are not supported.
func WrapAttrCtx(ctx context.Context, err error, meta ...slog.Attr) error {
	if err == nil {
		return nil
	}
	meta = appendFileToAttr(meta, err, caller-1, runtime.Frame{})
	meta = appendAttrFromCtx(ctx, meta)
	return maybeWrapAttrError(err, meta)
}

// WrapAttrCtxAfter is WrapAttrCtx for usage with defer.
// Defer at the top of a function with a named return error variable to wrap any error returned from the function with your desired metadata.
// An example of a function that returns structured errors to slog with convenience functions to reduce boilerplate:
//
//	func DeleteDevice(ctx context.Context, tx sqlx.ExecerContext, id uint64) (err error) {
//		defer errors.WrapAttrCtxAfter(ctx, &err, slog.Uint64("device_id", id))
//		propQuery := `DELETE FROM device_properties WHERE device_id = ?`
//		propsResult, err := tx.ExecContext(ctx, propQuery, id)
//		if err != nil {
//			return errors.Wrapf(err, "failed deleting device properties")
//		}
//
//		propsDeleted, err := propsResult.RowsAffected()
//		if err != nil {
//			return errors.Wrapf(err, "failed counting deleted device properties")
//		}
//		defer errors.WrapAttrCtxAfter(ctx, &err, slog.Uint64("deleted_props", propsDeleted))
//
//		query := `DELETE FROM device WHERE id = ?`
//		_,err = tx.ExecContext(ctx, query, id)
//		return errors.Wrapf(err, "tx.Exec failed")
//	}
//
// The output of slogging this function's failure with slog.Errorf("db error", "err", err):
// 2025/06/26 15:22:57 ERROR db error err.msg="errors.DeleteDevice tx.Exec failed" err.device_id=9 err.deleted_props=5 err.source=github.com/danlock/pkg/errors/ctx.go:76
//
// Using defer WrapAttrCtxAfter throughout our code makes it more maintainable by adding metadata when it's available, only if the error exists.
// Consider using WrapAttrCtxAfter after any error returning function with a context.Context parameter.
func WrapAttrCtxAfter(ctx context.Context, errPtr *error, meta ...slog.Attr) {
	if errPtr == nil {
		panic("WrapAttrCtxAfter errPtr must point at the caller function's named return error variable")
	}
	if *errPtr == nil {
		return
	}
	err := *errPtr
	meta = appendFileToAttr(meta, err, caller-1, runtime.Frame{})
	meta = appendAttrFromCtx(ctx, meta)
	*errPtr = maybeWrapAttrError(err, meta)
}
