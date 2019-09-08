// Copyright (c) 2019 Hervé Gouchet. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package backoff

const errRetryMsg = "maximum execution number exhausted"

// Common errors.
const (
	// ErrRetry is returns when the maximum execution number is exhausted.
	ErrRetry = errBackoff(errRetryMsg)
	// ErrDeadlineExceeded is the error returned when the context's deadline passes.
	ErrDeadlineExceeded = errBackoff("context deadline exceeded")
)

type errBackoff string

// Error implements the error interface.
func (e errBackoff) Error() string {
	return "backoff: " + string(e)
}

// newErrRetry embeds the error in a retry error.
func newErrRetry(err error) error {
	if err != nil {
		return errBackoff(errRetryMsg + ": " + err.Error())
	}
	return ErrRetry
}
