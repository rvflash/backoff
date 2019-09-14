// Copyright (c) 2019 Hervé Gouchet. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package backoff

import (
	"context"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestFibonacci(t *testing.T) {
	f := fibonacci()
	are := is.New(t)
	for _, v := range []time.Duration{1, 1, 2, 3, 5, 8, 13, 21, 34, 55} {
		are.Equal(f(), v) // mismatch suite
	}
}

func TestNew(t *testing.T) {
	bo := New(context.Background())
	is.New(t).Equal(bo.Attempt(), 0)
}

func TestBackoff_Do(t *testing.T) {
	// Fib check
	var bo Backoff
	n, err := bo.Do(void)
	are := is.New(t)
	are.Equal(err, context.Canceled) // mismatch error
	are.Equal(n, 0)                  // mismatch attempt
}

func TestBackoff_Retry(t *testing.T) {
	// No context
	var (
		bo  *Backoff
		ctx context.Context
		are = is.New(t)
	)
	bo = &Backoff{}
	n, err := bo.Retry(void)
	are.Equal(err, context.Canceled) // mismatch error
	are.Equal(n, 0)                  // mismatch attempt
	bo = New(ctx)
	n, err = bo.Retry(void)
	are.NoErr(err)  // unexpected error
	are.Equal(n, 0) // mismatch attempt
}

func TestBackoff_Reset(t *testing.T) {
	bo := New(context.Background()).WithInterval(time.Millisecond)
	are := is.New(t)
	are.True(bo.interval == time.Millisecond)
	bo.Reset()
	are.True(bo.interval == DefaultInterval)
}

// void implements the Func interface.
func void(context.Context) error {
	return nil
}
