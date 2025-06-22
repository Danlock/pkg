// Package retry provides retryable context-aware functions for code that needs to be robust against transient failures.
package retry

import (
	"context"
	"time"
)

// UntilDone repeatedly calls the provided function until the context finishes.
func UntilDone(ctx context.Context, fn func()) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			fn()
		}
	}
}

var fibonacciDurations = [...]time.Duration{
	0, time.Second, time.Second, 2 * time.Second, 3 * time.Second, 5 * time.Second,
	8 * time.Second, 13 * time.Second, 21 * time.Second, 34 * time.Second,
}

// FibonacciDelay provides a simple, default delay function that increases according to the Fibonacci sequence for 10 attempts.
func FibonacciDelay(attempt uint) time.Duration {
	if attempt < uint(len(fibonacciDurations)) {
		return fibonacciDurations[attempt]
	}
	return fibonacciDurations[len(fibonacciDurations)-1]
}

// WithBackoff repeatedly calls a function until the context finishes. The return value of the function is used to determine the backoff between retries.
// If the function returned true, the backoff is delay(0). If false, the backoff is delay(number of failed attempts).
// FibonacciDelay is used when delay is nil.
func WithBackoff(ctx context.Context, delay func(attempt uint) time.Duration, fn func() bool) {
	WithMaxAttempts(ctx, 0, delay, fn)
}

// WithMaxAttempts repeatedly calls a function until the context finishes. The return value of the function is used to determine the backoff between retries.
// If the function returned true, the backoff is delay(0). If false, the backoff is delay(number of failed attempts).
// FibonacciDelay is used when delay is nil.
// WithMaxAttempts also stops retrying after max attempt are reached as long as maxAttempts is greater than 0.
func WithMaxAttempts(ctx context.Context, maxAttempts uint, delay func(attempt uint) time.Duration, fn func() bool) {
	if delay == nil {
		delay = FibonacciDelay
	}

	var attempts uint
	tmr := time.NewTimer(0)
	defer tmr.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			select {
			case <-ctx.Done():
				return
			case <-tmr.C:
			}
		}

		if fn() {
			attempts = 0
		} else if maxAttempts > 0 && attempts >= maxAttempts {
			return
		} else {
			attempts++
		}

		tmr.Reset(delay(attempts))
	}
}
