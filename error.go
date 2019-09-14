// Copyright (c) 2019 Hervé Gouchet. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package backoff

const (
	errPrefix      = "backoff: "
	errRetryMsg    = "maximum execution number exhausted"
	errDeadlineMsg = "context deadline exceeded"
)

// Common errors.
var (
	// ErrRetry is returns when the maximum execution number is exhausted.
	ErrRetry = &Error{Reason: errRetryMsg}
	// ErrDeadlineExceeded is the error returned when the context's deadline passes.
	ErrDeadlineExceeded = &Error{Reason: errDeadlineMsg}
)

// Error represents an error.
type Error struct {
	Reason string
	Err    error
}

// DeadlineExceeded reports whether this error represents a deadline exceeded.
func (e Error) DeadlineExceeded() bool {
	return e.Reason == errDeadlineMsg
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.Err != nil {
		return errPrefix + e.Reason + ": " + e.Err.Error()
	}
	return errPrefix + e.Reason
}

// Unwrap returns the embedded error.
func (e Error) Unwrap(err error) error {
	return e.Err
}

// newErrRetry embeds the error in a retrying error.
func newErrRetry(err error) error {
	if err != nil {
		return &Error{Reason: errRetryMsg, Err: err}
	}
	return ErrRetry
}
