/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package task

// Task is an invocable unit of work
type Task interface {
	// Invoke invokes the task
	Invoke()

	// Attempts returns the number of invocation attempts that were made
	// in order to achieve a successful response
	Attempts() int

	// LastError returns the last error that occurred
	LastError() error
}
