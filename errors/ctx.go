package errors

import (
	"context"
	"log/slog"
	"runtime"
)

type ctxKey struct{}

// WrapMeta is WrapMetaCtx without the context.
func WrapMeta(err error, meta ...slog.Attr) error {
	if err == nil {
		return nil
	}
	meta = appendFileToMeta(meta, err, caller-1, runtime.Frame{})
	return maybeWrapMetaError(err, meta)
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

// maybeWrapMetaError wraps the error only if we have metadata or the error isn't a MetaError.
func maybeWrapMetaError(err error, meta []slog.Attr) error {
	// Wrapping the error with a metaError without metadata or a better error message
	// just bloats the linked list of errors without any benefit.
	//
	// However we do need a MetaError at the top of the error chain so our slog.LogValuer will be called.
	_, isMeta := err.(MetaError)
	if len(meta) == 0 && isMeta {
		return err
	}
	merr := metaError{error: err}
	merr.r.AddAttrs(meta...)
	return merr
}

// AddMetaToCtx adds metadata to the context that will be added to the error once WrapMetaCtx is called.
// It appends to any existing metadata in the context.
//
// The only way to retrieve the metadata is with UnwrapMeta on an error wrapped with WrapMetaCtx,
// or by just slogging the error which handles this internally.
//
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
//
// If not wrapping an error from this Go package it also includes the file and line info of it's caller.
// AddMetaToCtx adds metadata to the ctx which will be added to the error, if the context is set.
//
// If 0 metadata will be included with the error, i.e both context is nil and meta is empty,
// the original error will be returned to avoid bloating the error chain.
//
// Note that the slog output contains 2 keys by default, DefaultSourceSlogKey and DefaultMsgSlogKey,
// which use slog's standard "source" and "msg". Duplicate keys are not supported.
func WrapMetaCtx(ctx context.Context, err error, meta ...slog.Attr) error {
	if err == nil {
		return nil
	}
	meta = appendFileToMeta(meta, err, caller-1, runtime.Frame{})
	meta = appendMetaFromCtx(ctx, meta)
	return maybeWrapMetaError(err, meta)
}

// WrapMetaCtxAfter is WrapMetaCtx for usage with defer.
// Defer at the top of a function with a named return error variable to wrap any error returned from the function with your desired metadata.
// An example of a function that returns structured errors to slog with convenience functions to reduce boilerplate:
//
//	func DeleteDevice(ctx context.Context, tx sqlx.ExecerContext, id uint64) (res sql.Result, err error) {
//		defer errors.WrapMetaCtxAfter(ctx, &err, slog.Uint64("device_id", id))
//		propQuery := `DELETE FROM device_properties WHERE device_id = ?`
//		propsResult, err := tx.ExecContext(ctx, propQuery, id)
//		if err != nil {
//			return res, errors.Wrapf(err, "failed deleting device properties")
//		}
//
//		propsDeleted, err := propsResult.RowsAffected()
//		if err != nil {
//			return res, errors.Wrapf(err, "failed counting deleted device properties")
//		}
//		defer errors.WrapMetaCtxAfter(ctx, &err, slog.Uint64("deleted_props", propsDeleted))
//
//		query := `DELETE FROM device WHERE id = ?`
//		return errors.WrapfAndPass(tx.ExecContext(ctx, query, id))("tx.Exec failed")
//	}
//
// The output of slogging this function's failure with slog.Errorf("db error", "err", err):
// 2025/06/26 15:22:57 ERROR db error err.msg="errors.DeleteDevice tx.Exec failed" err.device_id=9 err.deleted_props=5 err.source=github.com/danlock/pkg/errors/ctx.go:76
//
// Using defer WrapMetaCtxAfter throughout our code makes it more maintainable by adding metadata when it's available, only if the error exists.
// Consider using WrapMetaCtxAfter after any error returning function with a context.Context parameter.
func WrapMetaCtxAfter(ctx context.Context, errPtr *error, meta ...slog.Attr) {
	if errPtr == nil {
		panic("WrapMetaCtxAfter errPtr must point at the caller function's named return error variable")
	}
	if *errPtr == nil {
		return
	}
	err := *errPtr
	meta = appendFileToMeta(meta, err, caller-1, runtime.Frame{})
	meta = appendMetaFromCtx(ctx, meta)
	*errPtr = maybeWrapMetaError(err, meta)
}
