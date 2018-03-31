/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invokeerror

import "github.com/pkg/errors"

type ErrorCode int

const (
	// PersistentError indicates that an unknown error occurred
	// and a retry of the request would NOT be successful
	PersistentError ErrorCode = iota

	// TransientError indicates that a transient error occurred
	// and a retry of the request could be successful
	TransientError

	// TimeoutOnCommit indicates that a timeout occurred while waiting
	// for a TxStatus event
	TimeoutOnCommit
)

// Error extends error with
// additional context
type Error interface {
	error

	// Status returns the error code
	ErrorCode() ErrorCode
}

type invokeError struct {
	error
	code ErrorCode
}

// New returns a new Error
func New(code ErrorCode, msg string) Error {
	return &invokeError{
		error: errors.New(msg),
		code:  code,
	}
}

// Wrap returns a new Error
func Wrap(code ErrorCode, cause error, msg string) Error {
	return &invokeError{
		error: errors.Wrap(cause, msg),
		code:  code,
	}
}

// Wrapf returns a new Error
func Wrapf(code ErrorCode, cause error, fmt string, args ...interface{}) Error {
	return &invokeError{
		error: errors.Wrapf(cause, fmt, args...),
		code:  code,
	}
}

// Errorf returns a new Error
func Errorf(code ErrorCode, fmt string, args ...interface{}) Error {
	return &invokeError{
		error: errors.Errorf(fmt, args...),
		code:  code,
	}
}

// ErrorCode returns the error code
func (e *invokeError) ErrorCode() ErrorCode {
	return e.code
}
