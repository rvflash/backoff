# Fibonacci backoff implementation

[![GoDoc](https://godoc.org/github.com/rvflash/backoff?status.svg)](https://godoc.org/github.com/rvflash/backoff)
[![Build Status](https://img.shields.io/travis/rvflash/backoff.svg)](https://travis-ci.org/rvflash/backoff)
[![Code Coverage](https://img.shields.io/codecov/c/github/rvflash/backoff.svg)](http://codecov.io/github/rvflash/backoff?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/rvflash/backoff)](https://goreportcard.com/report/github.com/rvflash/backoff)

The package `backoff` implements a Fibonacci backoff algorithm.


### Installation
    
To install it, you need to install Go and set your Go workspace first.
Then, download and install it:

```bash
$ go get -u github.com/rvflash/backoff
```    
Import it in your code:
    
```go
import "github.com/rvflash/backoff"
```

### Prerequisite

`backoff` uses the Go modules that required Go 1.11 or later.


## Features

Based on the Fibonacci suite (1, 1, 2, 3, 5, 8, 13, 21, etc.), the `backoff` strategy do or retry a given task.
By default, `DefaultInterval` is used as interval, so the sleep duration is the `current Fibonacci value * 500 * time.Millisecond`.
With the `New` method, you can create your own Backoff strategy but by default, the following implementation are available:  
See the documentation for more details and samples.


### Do 

`Do` guarantees to execute at least once the task if the context is not already cancelled.
As long as the task return in success and the context not done, BackOff will continue to call it, with a sleep duration based the Fibonacci suite and the BackOff's interval.

* DoN: does the same job as Do but limits the number of attempt.
* DoUntil: does the same job as Do but limits the execution to the given deadline.


### Retry 

`Retry` retries the task until it does not return error or BackOff stops.

* RetryN: does the same job as Do but limits the number of attempt.
* RetryUntil: does the same job as Do but limits the execution to the given deadline.


## Quick start

Assuming the following code that retry 3 times to run the task if it returns in error.

```go

import (
	"context"
	"log"

	"github.com/rvflash/backoff"
)

func main() {
	// task implements the backoff.Func interface.
	task := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return nil
	}
	n, err := backoff.RetryN(context.Background(), 3, task)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Nice boy: %d retry, first try in success", n)
	// Output: Nice boy: 0 retry, first try in success
}
```