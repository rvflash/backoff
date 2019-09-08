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