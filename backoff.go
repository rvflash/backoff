// Copyright (c) 2019 Hervé Gouchet. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

// Package backoff provides a Fibonacci backoff implementation.
package backoff

import (
	"context"
	"sync"
	"time"
)

const (
	// DefaultInterval is the default interval between 2 iterations.
	DefaultInterval = 500 * time.Millisecond
)

type funcAlgorithm func() time.Duration

// fibonacci implements the Fibonacci suite.
func fibonacci() funcAlgorithm {
	var (
		a time.Duration
		b time.Duration = 1
	)
	return func() time.Duration {
		a, b = b, a+b
		return a
	}
}

// Func must be implemented by any function to be run by the Backoff.
type Func func(context.Context) error

// Do guarantees to execute at least once f if ctx is not already cancelled.
// As long as f return in success and the context not done, BackOff will continue to call it,
// with sleep duration based the Fibonacci suite and the BackOff's interval.
func Do(ctx context.Context, f Func) (int, error) {
	return New(ctx).Do(f)
}

// DoN does the same job as Do but limits the number of attempt to n.
func DoN(ctx context.Context, attempt int, f Func) (int, error) {
	return New(ctx).WithMaxAttempt(attempt).Do(f)
}

// DoUntil does the same job as Do but limits ctx by adjusting the deadline to be no later than d.
func DoUntil(ctx context.Context, t time.Time, f Func) (int, error) {
	return New(ctx).WithDeadline(t).Do(f)
}

// Retry retries the function f until it does not return error or BackOff stops.
// f is guaranteed to be run at least once, unless the context is already cancelled.
func Retry(ctx context.Context, f Func) (int, error) {
	return New(ctx).Retry(f)
}

// RetryN does the same as Retry but limits the number of attempt to n.
func RetryN(ctx context.Context, attempt int, f Func) (int, error) {
	return New(ctx).WithMaxAttempt(attempt).Retry(f)
}

// RetryUntil does the same job as Retry but limits ctx by adjusting the deadline to be no later than d.
func RetryUntil(ctx context.Context, t time.Time, f Func) (int, error) {
	return New(ctx).WithDeadline(t).Retry(f)
}

// Retryer is a Backoff strategy for retrying an operation based on the Fibonacci suite.
type Retryer interface {
	// Attempt returns the current number of attempt.
	Attempt() int
	// Do executes the given function every "fib tick" as long as it is successful.
	// A context cancelled, a deadline or maximum attempt exceeded can also break the loop.
	Do(f Func) (int, error)
	// Reset resets to initial state.
	Reset()
	// Retry executes the given function every "fib tick" as long as it is failed.
	// A context cancelled, a deadline or maximum attempt exceeded can also break the loop.
	Retry(f Func) (int, error)
	// WithDeadline creates a copy of the current Backoff to defines a new context
	// with the deadline adjusted to be no later than t.
	WithDeadline(t time.Time) Retryer
	// WithInterval sets the time interval between two try with the value of d.
	WithInterval(d time.Duration) Retryer
	// WithMaxAttempt sets the maximum number of attempt to n.
	WithMaxAttempt(n int) Retryer
}

// New returns a new instance of Backoff.
func New(ctx context.Context) *Backoff {
	if ctx == nil {
		ctx = context.Background()
	}
	b := newBackoff()
	b.ctx, b.cancel = context.WithCancel(ctx)
	return b
}

func newBackoff() *Backoff {
	return &Backoff{
		interval: DefaultInterval,
		err:      make(chan error),
		fib:      fibonacci(),
	}
}

// Backoff is a time.Duration and an attempt counter.
// It provides means to do and retry something based on the Fibonacci suite as trigger.
type Backoff struct {
	ctx    context.Context
	cancel context.CancelFunc
	err    chan error
	fib    funcAlgorithm

	attempt,
	maxAttempt int
	interval time.Duration
	mu       sync.RWMutex
}

// Attempt implements the Retryer interface.
func (b *Backoff) Attempt() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.attempt
}

// Do implements the Retryer interface.
func (b *Backoff) Do(f Func) (int, error) {
	if b.fib == nil {
		return b.Attempt(), context.Canceled
	}
	go b.run(f, false)
	return b.done()
}

// Reset implements the Retryer interface.
func (b *Backoff) Reset() {
	b.mu.Lock()
	b.attempt = 0
	b.fib = fibonacci()
	b.mu.Unlock()
}

// Retry implements the Retryer interface.
func (b *Backoff) Retry(f Func) (int, error) {
	if b.fib == nil {
		return b.Attempt(), context.Canceled
	}
	go b.run(f, true)
	return b.done()
}

// WithDeadline implements the Retryer interface.
func (b *Backoff) WithDeadline(t time.Time) Retryer {
	b2 := b.copy()
	b2.ctx, b2.cancel = context.WithDeadline(b.ctx, t)
	return b2
}

// WithInterval implements the Retryer interface.
func (b *Backoff) WithInterval(d time.Duration) Retryer {
	if d > 0 {
		b.mu.Lock()
		b.interval = d
		b.mu.Unlock()
	}
	return b
}

// WithMaxAttempt implements the Retryer interface.
func (b *Backoff) WithMaxAttempt(n int) Retryer {
	if n > -1 {
		b.mu.Lock()
		b.maxAttempt = n
		b.mu.Unlock()
	}
	return b
}

// copy copies the Backoff to create a new one with the same behavior.
// It also takes care of the underlying mutex.
func (b *Backoff) copy() *Backoff {
	b2 := newBackoff()
	b.mu.Lock()
	b2.interval = b.interval
	b2.maxAttempt = b.maxAttempt
	b.mu.Unlock()
	return b2
}

// done waits the end of the job, done or cancelled.
func (b *Backoff) done() (int, error) {
	defer b.cancel()
	select {
	case <-b.ctx.Done():
		return b.Attempt(), ErrDeadlineExceeded
	case err := <-b.err:
		return b.Attempt(), err
	}
}

// next increments the number of attempt, to validate or not the go to the next iteration.
func (b *Backoff) next() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.attempt++
	if b.maxAttempt > 0 && b.attempt >= b.maxAttempt {
		// Maximum number of attempt exceeded.
		return ErrRetry
	}
	return nil
}

// run runs the Retryer strategy by using f as job to do and retry as mode.
func (b *Backoff) run(f Func, retry bool) {
	var err, rrr error
	for {
		select {
		case <-b.ctx.Done():
			// Context cancelled.
			return
		default:
		}
		err = f(b.ctx)
		switch {
		case
			// Do is finished when an error has occurred.
			!retry && err != nil,
			// Retry is finished when no error occurred.
			retry && err == nil:
			// Job done.
			b.err <- err
			return
		}
		// Tries to begin a new iteration.
		rrr = b.next()
		if rrr != nil {
			b.err <- newErrRetry(err)
			return
		}
		// Waiting before to run the next iteration.
		rrr = b.sleep()
		if rrr != nil {
			b.err <- newErrRetry(err)
			return
		}
	}
}

// sleep pauses the current goroutine for at least the duration of the interval
// multiplied by the current Fibonacci value.
func (b *Backoff) sleep() error {
	b.mu.Lock()
	d := b.fib() * b.interval
	b.mu.Unlock()
	if d < 0 {
		// Bound exceeded.
		return ErrRetry
	}
	time.Sleep(d)
	return nil
}
