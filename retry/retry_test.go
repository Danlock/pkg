package retry

import (
	"context"
	"testing"
	"time"

	"github.com/danlock/pkg/test"
)

func TestUntilDone(t *testing.T) {
	ctx, _ := context.WithTimeout(t.Context(), 10*time.Millisecond)

	count := 0

	go UntilDone(ctx, func() {
		count++
		ctx, _ := context.WithTimeout(ctx, time.Millisecond)
		<-ctx.Done()
	})

	<-ctx.Done()
	test.Equality(t, 10, count, "unexpected count == %d", count)
}

func TestWithMaxAttempts(t *testing.T) {
	ctx, _ := context.WithTimeout(t.Context(), 10*time.Millisecond)
	count := 0

	go WithMaxAttempts(ctx, 0, func(uint) time.Duration { return 0 }, func() bool {
		count++
		ctx, _ := context.WithTimeout(ctx, 3*time.Millisecond)
		<-ctx.Done()
		return true
	})

	<-ctx.Done()
	test.Equality(t, 4, count, "unexpected count == %d", count)

	count = 0
	ctx, _ = context.WithTimeout(t.Context(), 10*time.Millisecond)

	go WithMaxAttempts(ctx, 1, nil, func() bool {
		count++
		ctx, _ := context.WithTimeout(ctx, 3*time.Millisecond)
		<-ctx.Done()
		return false
	})

	<-ctx.Done()
	test.Equality(t, 1, count, "unexpected count == %d", count)
}
