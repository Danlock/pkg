package errors

import (
	"fmt"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/danlock/pkg/test"
)

func TestMeta(t *testing.T) {
	attr1 := slog.String("key", "value")
	attr2 := slog.Uint64("id", 1234)
	attr3 := slog.Time("ts", time.Time{})

	oops := func() error {
		return WrapMeta(New("oops"), attr1, attr2)
	}

	regErr := func(err error) error {
		return fmt.Errorf("stdlib %w", err)
	}
	var err error = error(nil)
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
			nil,
			"errors.TestMeta.func1 oops",
			"errors.TestMeta.func1 oops {key=value,id=1234}",
		},
		{
			"triple decker",
			func() error {
				return WrapMeta(regErr(oops()), attr3)
			}(),
			true,
			nil,
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
