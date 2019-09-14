// Copyright (c) 2019 Hervé Gouchet. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package backoff_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/rvflash/backoff"
)

func newTask(until int) *task {
	return &task{start: time.Now(), until: until}
}

type task struct {
	// called contains the number of iteration done
	called,
	// until contains the number of iteration to do in success
	until int
	// start is the start time
	start time.Time
	// uptime return the time since the last call
	latest time.Duration
}

var errRetry = errors.New("oops")

func (t *task) stopwatch() {
	t.latest = time.Since(t.start)
}

// KoUntil implements the backoff.Func interface.
func (t *task) KoUntil(context.Context) error {
	defer t.stopwatch()
	t.called++
	if t.called < t.until {
		return errRetry
	}
	return nil
}

// OkUntil implements the backoff.Func interface.
func (t *task) OkUntil(context.Context) error {
	defer t.stopwatch()
	t.called++
	if t.called > t.until {
		return errRetry
	}
	return nil
}

func TestDo(t *testing.T) {
	var (
		are = is.New(t)
		dt  = map[string]struct {
			f   backoff.Func
			n   int
			err error
		}{
			"ko: no ctx: 1 times":        {f: newTask(0).OkUntil, n: 0, err: errRetry},
			"ko: no ctx: 2 times":        {f: newTask(1).OkUntil, n: 1, err: errRetry},
			"ko: ctx cancelled: 2 times": {f: newTask(4).OkUntil, n: 3, err: backoff.ErrDeadlineExceeded},
		}
	)
	for name, tt := range dt {
		tt := tt
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			n, err := backoff.Do(ctx, tt.f)
			cancel()
			are.Equal(n, tt.n)     // mismatch attempt
			are.Equal(err, tt.err) // mismatch error
		})
	}
}

func TestDoN(t *testing.T) {
	const n = 3
	var (
		are = is.New(t)
		job = newTask(5)
	)
	i, err := backoff.DoN(context.Background(), n, job.OkUntil)
	are.Equal(err, backoff.ErrRetry) // mismatch error
	are.Equal(i, n)                  // mismatch attempt
	are.Equal(i, job.called)         // mismatch call
}

func TestDoUntil(t *testing.T) {
	const d = time.Second
	var (
		are = is.New(t)
		dur = time.Now().Add(d)
		job = newTask(5)
	)
	i, err := backoff.DoUntil(context.Background(), dur, job.OkUntil)
	are.Equal(err, backoff.ErrDeadlineExceeded) // mismatch error
	are.Equal(i, 2)                             // mismatch attempt
	are.True(job.latest < d)                    // mismatch duration
}

func TestRetry(t *testing.T) {
	var (
		are = is.New(t)
		dt  = map[string]struct {
			f   backoff.Func
			n   int
			err error
		}{
			"ok: no retry":               {f: newTask(1).KoUntil},
			"ok: 1 times":                {f: newTask(2).KoUntil, n: 1},
			"ko: ctx cancelled: 2 times": {f: newTask(5).KoUntil, n: 3, err: backoff.ErrDeadlineExceeded},
		}
	)
	for name, tt := range dt {
		tt := tt
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			n, err := backoff.Retry(ctx, tt.f)
			cancel()
			are.Equal(n, tt.n)     // mismatch attempt
			are.Equal(err, tt.err) // mismatch error
		})
	}
}

func TestRetryN(t *testing.T) {
	const n = 3
	var (
		are = is.New(t)
		job = newTask(5)
	)
	var backoffError *backoff.Error
	i, err := backoff.RetryN(context.Background(), n, job.KoUntil)
	are.True(errors.As(err, &backoffError))    // mismatch error
	are.True(!backoffError.DeadlineExceeded()) // mismatch type
	are.Equal(i, n)                            // mismatch attempt
	are.Equal(i, job.called)                   // mismatch call
}

func TestRetryUntil(t *testing.T) {
	const (
		n = 2
		d = time.Second
	)
	var (
		are = is.New(t)
		job = newTask(5)
	)
	var backoffError *backoff.Error
	i, err := backoff.RetryUntil(context.Background(), time.Now().Add(d), job.KoUntil)
	are.True(errors.As(err, &backoffError))   // mismatch error
	are.True(backoffError.DeadlineExceeded()) // mismatch type
	are.Equal(i, n)                           // mismatch attempt
	are.True(job.latest < d)                  // mismatch duration
}
