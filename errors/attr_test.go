package errors

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/danlock/pkg/test"
)

func setup() {
	// This is just setup code that makes slog's output deterministic so the example output is stable.
	DefaultSourceSlogKey = slog.SourceKey
	ShouldSortAttr = true
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})))
}

func baby() error { return New("don't hurt me") }

func Example() {
	setup()
	// This example showcases how to use structured errors alongside log/slog.
	err := baby()
	if err != nil {
		// include some metadata about this failure
		err = WrapAttr(err, slog.String("don't", "hurt me"), slog.String("no", "more"))
	}
	// Typically this error would then bubble up through a few more function calls.
	// Could be wrapped many more times, but eventually something handles this error.
	// For exanple, it can be logged
	if err != nil {
		slog.Warn("what is love", "err", err)
	}
	// Pulling out metadata from a context is also possible, useful for attaching something like request IDs to any error from a request handler.
	ctx := AddAttrToCtx(context.Background(), slog.Uint64("answer", 42))
	// WrapAttrCtxAfter is an simple and maintainable way to add context metadata to any error returned from a function.
	// Here is a small function that hashes and writes some random bytes to showcase various error helper functions from this package.
	_, err = func(ctx context.Context, file string) (_ int, err error) {
		dest := path.Join(os.TempDir(), "hashed.brown")
		defer WrapAttrCtxAfter(ctx, &err, slog.String("input", file), slog.String("output", dest))
		fileBytes := make([]byte, 10)
		// Scrounge up some bytes
		bytesRead, err := rand.NewChaCha8([32]byte{}).Read(fileBytes)
		if err != nil {
			return 0, Wrapf(err, "failed to generate bytes")
		}
		// Ensure we track how much we read in case that's relevant later
		defer WrapAttrCtxAfter(ctx, &err, slog.Int("bytes_read", bytesRead))

		hash := sha256.Sum256(fileBytes)
		// Open this file for writing... or reading... whatever.
		f, err := os.OpenFile(path.Clean(dest), os.O_RDONLY, 0600)
		if err != nil {
			return 0, Wrapf(err, "failed os.OpenFile as read only")
		}
		// JoinAfter helps you respect errors from commonly ignored functions like Close.
		defer JoinAfter(&err, f.Close)
		// If you're familiar with github.com/pkg/errors, you may be used to ending error returning functions with `return errors.Wrap(err)`
		// WrapfAndPass extends that to functions returning a value and an error.
		return WrapfAndPass(f.Write(hash[:]))("failed os.WriteFile")
	}(ctx, path.Join(os.TempDir(), "hash.brown"))

	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "hash browns burnt", slog.Any("err", err))
	}

	// printing the error with something like fmt.Println won't include the metadata in the output.
	fmt.Println(err)
	err = Wrapf(err, "doubleWrap")
	// unless you use %+v
	fmt.Printf("%+v", err)

	// Output:
	// level=WARN msg="what is love" err.msg="errors.baby don't hurt me" err.don't="hurt me" err.no=more err.source=github.com/danlock/pkg/errors/attr_test.go:32
	// level=ERROR msg="hash browns burnt" err.msg="errors.Example.func1 failed os.WriteFile write /tmp/hashed.brown: bad file descriptor" err.answer=42 err.bytes_read=10 err.input=/tmp/hash.brown err.output=/tmp/hashed.brown err.source=github.com/danlock/pkg/errors/attr_test.go:74
	// errors.Example.func1 failed os.WriteFile write /tmp/hashed.brown: bad file descriptor
	// [msg=errors.Example doubleWrap errors.Example.func1 failed os.WriteFile write /tmp/hashed.brown: bad file descriptor answer=42 bytes_read=10 input=/tmp/hash.brown output=/tmp/hashed.brown source=github.com/danlock/pkg/errors/attr_test.go:74]
}

func TestAttr(t *testing.T) {
	attr1 := slog.String("key", "value")
	attr2 := slog.Uint64("id", 1234)
	attr3 := slog.Time("ts", time.Time{})
	attr4 := slog.Bool("bit", true)

	DefaultSourceSlogKey = ""

	oops := func() error {
		return WrapAttr(New("oops"), attr1, attr2)
	}

	regErr := func(err error) error {
		return fmt.Errorf("stdlib %w", err)
	}

	test.Equality(t, slog.KindString, UnwrapAttr(oops())[attr1.Key].Kind())

	var err = error(nil)
	tests := []struct {
		name      string
		err       error
		wantErr   bool
		wantAttr  []slog.Attr
		wantBasic string
	}{
		{
			"zilch",
			WrapAttr(err, attr1, attr2),
			false,
			nil,
			"",
		},
		{
			"single layer",
			oops(),
			true,
			[]slog.Attr{attr1, attr2},
			"errors.TestAttr.func1 oops",
		},
		{
			"triple decker",
			func() error {
				return WrapAttr(regErr(oops()), attr3)
			}(),
			true,
			[]slog.Attr{attr3, attr1, attr2},
			"stdlib errors.TestAttr.func1 oops",
		},
		{
			"the fat bastard",
			func() error {
				return Wrap(Join(Wrap(Join(WrapAttr(nil), WrapAttr(regErr(oops()), attr3), Wrap(nil), WrapAttr(New("please stop"), attr4))), WrapAttr(New("No dupes"), attr1)))
			}(),
			true,
			[]slog.Attr{attr3, attr1, attr2, attr4},
			"errors.TestAttr.func4 errors.TestAttr.func4 stdlib errors.TestAttr.func1 oops\nerrors.TestAttr.func4 please stop\nerrors.TestAttr.func4 No dupes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.Equality(t, tt.wantErr, tt.err != nil, "WrapAttr() error = %+v, wantErr %v", tt.err, tt.wantErr)

			if len(tt.wantAttr) > 0 {
				metaMap := UnwrapAttr(tt.err)
				expandedStr := fmt.Sprintf("%+v", tt.err)
				for _, attr := range tt.wantAttr {
					attrStr := attr.String()
					test.Truth(t, strings.Contains(expandedStr, attrStr), "expanded error string %s didn't contain %s", expandedStr, attrStr)

					v, ok := metaMap[attr.Key]
					test.Truth(t, ok && v.Equal(attr.Value), "err metadata %+v missing attr %s", metaMap, attrStr)
				}
			}
			if len(tt.wantBasic) > 0 {
				test.Equality(t, tt.wantBasic, fmt.Sprintf("%v", tt.err), "fmt.Sprintf %%v")
				test.Equality(t, tt.wantBasic, fmt.Sprintf("%s", tt.err), "fmt.Sprintf %%s")
				test.Equality(t, tt.wantBasic, fmt.Sprint(tt.err), "fmt.Sprint")
			}
		})
	}
}
