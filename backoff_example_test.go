// Copyright (c) 2019 Hervé Gouchet. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package backoff_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rvflash/backoff"
)

func ExampleRetry() {
	n, err := backoff.Retry(context.Background(), Task)
	if err != nil {
		fmt.Println("err:", n)
	}
	fmt.Println("n:", n)
	// Output: Job done.
	// n: 0
}

func ExampleRetryN() {
	n, err := backoff.RetryN(context.Background(), 3, TaskInErr)
	if err != nil {
		fmt.Println("err:", err)
	}
	fmt.Println("n:", n)
	// Output: err: backoff: maximum execution number exhausted: oops
	// n: 3
}

func ExampleRetryUntil() {
	n, err := backoff.RetryUntil(context.Background(), time.Now().Add(time.Second), TaskInErr)
	if err != nil {
		fmt.Println("err:", err)
	}
	fmt.Println("n:", n)
	// Output: err: backoff: context deadline exceeded
	// n: 2
}

// Task implements the backoff.Func interface.
func Task(context.Context) error {
	fmt.Println("Job done.")
	return nil
}

var errTask = errors.New("oops")

// Task implements the backoff.Func interface.
func TaskInErr(context.Context) error {
	return errTask
}
