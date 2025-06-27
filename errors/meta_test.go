package errors

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/danlock/pkg/test"
)

func setup() {
	// This is just setup code that makes slog's output deterministic so the example output is stable.
	DefaultFileSlogKey = "file"
	ShouldSortUnwrapMeta = true
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})))
}

func dontHurtMe() error { return New("no more") }

func Example() {
	setup()
	// This example shows how to use WrapMeta() to attach metadata to errors.
	err := dontHurtMe()
	if err != nil {
		// include some metadata about this failure
		err = WrapMeta(err, slog.String("baby", "don't"), slog.String("hurt", "me"))
	}
	// Typically this error would then bubble up through a few more function calls.
	// Could be wrapped many more times, but eventually something handles this error.
	// For exanple, it can be logged
	if err != nil {
		slog.Warn("what is love", "err", err)
	}

	// Pulling out metadata from a context is also possible, useful for attaching something like request IDs to any error from a request handler.
	ctx := AddMetaToCtx(context.Background(), slog.Uint64("req_id", 42))
	// WrapMetaCtxAfter is an simple and maintainable way to add context metadata to any error returned from a function.
	// Wrap should be called as close to the error generating function as possible for accurate file and line info though.

	err = func(id uint64, parseMe string) (err error) {
		defer WrapMetaCtxAfter(ctx, &err, slog.Uint64("user_id", id))
		_, err = strconv.Atoi(parseMe)
		return Wrap(err)
	}(0451, "trust me i'm numerical")

	if err != nil {
		slog.LogAttrs(context.TODO(), slog.LevelWarn, "parse failure", slog.Any("err", err))
	}

	// printing the error with something like fmt.Println won't include the metadata in the output.
	fmt.Println(err)
	err = Wrapf(err, "doubleWrap")
	// unless you use %+v
	fmt.Printf("%+v", err)

	// Output: level=WARN msg="what is love" err.baby=don't err.file=github.com/danlock/pkg/errors/meta_test.go:30 err.hurt=me err.msg="errors.dontHurtMe no more"
	// level=WARN msg="parse failure" err.file=github.com/danlock/pkg/errors/meta_test.go:55 err.req_id=42 err.user_id=297 err.msg="errors.Example.func1 strconv.Atoi: parsing \"trust me i'm numerical\": invalid syntax"
	// errors.Example.func1 strconv.Atoi: parsing "trust me i'm numerical": invalid syntax
	// errors.Example doubleWrap errors.Example.func1 strconv.Atoi: parsing "trust me i'm numerical": invalid syntax {file=github.com/danlock/pkg/errors/meta_test.go:55,req_id=42,user_id=297}
}

func TestMeta(t *testing.T) {
	attr1 := slog.String("key", "value")
	attr2 := slog.Uint64("id", 1234)
	attr3 := slog.Time("ts", time.Time{})
	attr4 := slog.Bool("bit", true)

	DefaultFileSlogKey = ""

	oops := func() error {
		return WrapMeta(New("oops"), attr1, attr2)
	}

	regErr := func(err error) error {
		return fmt.Errorf("stdlib %w", err)
	}

	test.Equality(t, slog.KindString, UnwrapMetaMap(oops())[attr1.Key].Kind())

	var err = error(nil)
	tests := []struct {
		name      string
		err       error
		wantErr   bool
		wantMeta  []slog.Attr
		wantBasic string
	}{
		{
			"zilch",
			WrapMeta(err, attr1, attr2),
			false,
			nil,
			"",
		},
		{
			"single layer",
			oops(),
			true,
			[]slog.Attr{attr1, attr2},
			"errors.TestMeta.func1 oops",
		},
		{
			"triple decker",
			func() error {
				return WrapMeta(regErr(oops()), attr3)
			}(),
			true,
			[]slog.Attr{attr3, attr1, attr2},
			"stdlib errors.TestMeta.func1 oops",
		},
		{
			"the fat bastard",
			func() error {
				return Wrap(Join(Wrap(Join(WrapMeta(nil), WrapMeta(regErr(oops()), attr3), Wrap(nil), WrapMeta(New("please stop"), attr4))), WrapMeta(New("No dupes"), attr1)))
			}(),
			true,
			[]slog.Attr{attr3, attr1, attr2, attr4},
			"errors.TestMeta.func4 errors.TestMeta.func4 stdlib errors.TestMeta.func1 oops\nerrors.TestMeta.func4 please stop\nerrors.TestMeta.func4 No dupes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.Equality(t, tt.wantErr, tt.err != nil, "WrapMeta() error = %+v, wantErr %v", tt.err, tt.wantErr)

			if len(tt.wantMeta) > 0 {
				metaMap := UnwrapMetaMap(tt.err)
				expandedStr := fmt.Sprintf("%+v", tt.err)
				for _, attr := range tt.wantMeta {
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
