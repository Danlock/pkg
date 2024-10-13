package retry

import (
	"context"
	"testing"
	"time"
)

func TestUntilDone(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)

	count := 0

	go UntilDone(ctx, func() {
		count++
		ctx, _ := context.WithTimeout(ctx, time.Millisecond)
		<-ctx.Done()
	})

	<-ctx.Done()
	if count < 10 || count > 10 {
		t.Fatalf("unexpected count == %d", count)
	}
}

func TestWithMaxAttempts(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)

	count := 0

	go WithMaxAttempts(ctx, 0, func(attempt uint) time.Duration { return 0 }, func() bool {
		count++
		ctx, _ := context.WithTimeout(ctx, time.Millisecond)
		<-ctx.Done()
		return true
	})

	<-ctx.Done()
	if count < 10 || count > 10 {
		t.Fatalf("unexpected count == %d", count)
	}

	count = 0
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Millisecond)

	go WithMaxAttempts(ctx, 1, nil, func() bool {
		count++
		ctx, _ := context.WithTimeout(ctx, time.Millisecond)
		<-ctx.Done()
		return false
	})

	<-ctx.Done()
	if count < 1 || count > 1 {
		t.Fatalf("unexpected count == %d", count)
	}
}
