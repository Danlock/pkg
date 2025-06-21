package errors

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/danlock/pkg/test"
)

func dontHurtMe() error { return New("no more") }

func Example() {
	// This is just setup code that makes slog's output deterministic so the example output is stable.
	DefaultFileSlogKey = "file"
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})))
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
	// Another easy way of wrapping errors with metadata known at the start of the function is to defer WrapMeta.
	// This is possible since WrapMeta returns nil if the error is nil.

	err = func(id uint64, parseMe string) (err error) {
		defer func() { err = WrapMetaCtx(ctx, err, slog.Uint64("id", id)) }()
		_, err = strconv.Atoi(parseMe)
		if err != nil {
			return Wrap(err)
		}
		return nil
	}(0451, "trust me i'm numerical")

	if err != nil {
		slog.LogAttrs(context.TODO(), slog.LevelWarn, "parse failure", slog.Any("err", err))
	}

	// printing the error with something like fmt.Println won't include the metadata in the output.
	fmt.Println(err)
	err = Wrapf(err, "doubleWrap")
	// unless you use %+v
	fmt.Printf("%+v", err)

	// Output: level=WARN msg="what is love" err.baby=don't err.hurt=me err.file=/home/dan/go/src/github.com/danlock/pkg/errors/meta_test.go:16 err.msg="errors.dontHurtMe no more"
	// level=WARN msg="parse failure" err.id=297 err.req_id=42 err.file=/home/dan/go/src/github.com/danlock/pkg/errors/meta_test.go:51 err.msg="errors.Example.func2 strconv.Atoi: parsing \"trust me i'm numerical\": invalid syntax"
	// errors.Example.func2 strconv.Atoi: parsing "trust me i'm numerical": invalid syntax
	// errors.Example doubleWrap errors.Example.func2 strconv.Atoi: parsing "trust me i'm numerical": invalid syntax {id=297,req_id=42,file=/home/dan/go/src/github.com/danlock/pkg/errors/meta_test.go:51}
}

func TestMeta(t *testing.T) {
	attr1 := slog.String("key", "value")
	attr2 := slog.Uint64("id", 1234)
	attr3 := slog.Time("ts", time.Time{})

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
		name       string
		err        error
		wantErr    bool
		wantMeta   []slog.Attr
		wantBasic  string
		wantExpand string
	}{
		{
			"zilch",
			WrapMeta(err, attr1, attr2),
			false,
			nil,
			"",
			"",
		},
		{
			"single layer",
			oops(),
			true,
			[]slog.Attr{attr1, attr2},
			"errors.TestMeta.func1 oops",
			"errors.TestMeta.func1 oops {key=value,id=1234}",
		},
		{
			"triple decker",
			func() error {
				return WrapMeta(regErr(oops()), attr3)
			}(),
			true,
			[]slog.Attr{attr3, attr1, attr2},
			"stdlib errors.TestMeta.func1 oops",
			"stdlib errors.TestMeta.func1 oops {ts=0001-01-01 00:00:00 +0000 UTC,key=value,id=1234}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.Equality(t, tt.wantErr, tt.err != nil, "WrapMeta() error = %+v, wantErr %v", tt.err, tt.wantErr)

			if len(tt.wantMeta) > 0 && !reflect.DeepEqual(UnwrapMeta(tt.err), tt.wantMeta) {
				t.Errorf("UnwrapMeta() got = %+v, wanted %+v", UnwrapMeta(tt.err), tt.wantMeta)
			}
			if len(tt.wantBasic) > 0 {
				test.Equality(t, tt.wantBasic, fmt.Sprintf("%v", tt.err), "fmt.Sprintf %%v")
				test.Equality(t, tt.wantBasic, fmt.Sprintf("%s", tt.err), "fmt.Sprintf %%s")
				test.Equality(t, tt.wantBasic, fmt.Sprint(tt.err), "fmt.Sprint")
			}
			if len(tt.wantExpand) > 0 {
				test.Equality(t, tt.wantExpand, fmt.Sprintf("%+v", tt.err), "fmt.Sprintf %%+v")
			}
		})
	}
}
