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
	return &task{until: until}
}

type task struct {
	// called contains the number of iteration done
	called,
	// until contains the number of iteration to do in success
	until int
}

var errRetry = errors.New("oops")

// Run implements the backoff.Func interface.
func (t *task) Run(context.Context) error {
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
			"ko: no ctx: 1 times":        {f: newTask(0).Run, n: 0, err: errRetry},
			"ko: no ctx: 2 times":        {f: newTask(1).Run, n: 1, err: errRetry},
			"ko: ctx cancelled: 2 times": {f: newTask(4).Run, n: 3, err: backoff.ErrDeadlineExceeded},
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

func TestRetry(t *testing.T) {
	var (
		are = is.New(t)
		dt  = map[string]struct {
			f   backoff.Func
			n   int
			err error
		}{
			"ok: 1 times": {f: newTask(1).Run, n: 0},
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
