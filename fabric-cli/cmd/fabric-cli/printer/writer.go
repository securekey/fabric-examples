/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"fmt"
	"os"
	"strings"

	config "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
)

// WriterType specifies the format for printing data
type WriterType uint8

const (
	// STDOUT writes to standard out
	STDOUT WriterType = iota

	// STDERR writes to standard error
	STDERR

	// LOG writes to the logger
	LOG
)

const (
	stdout = "stdout"
	stderr = "stderr"
	log    = "log"
)

func (f WriterType) String() string {
	switch f {
	case STDOUT:
		return stdout
	case STDERR:
		return stderr
	case LOG:
		return log
	default:
		return "unknown"
	}
}

// AsWriterType returns the WriterType given a Writer Type string
func AsWriterType(t string) WriterType {
	switch strings.ToLower(t) {
	case log:
		return LOG
	case stderr:
		return STDERR
	default:
		return STDOUT
	}
}

// Writer writes the output
type Writer interface {
	Write(format string, a ...interface{}) error
}

// NewWriter returns a new writer given the writer type
func NewWriter(writerType WriterType) Writer {
	switch writerType {
	case STDERR:
		return &stdErrWriter{}
	case LOG:
		return &logWriter{}
	default:
		return &stdOutWriter{}
	}
}

type stdOutWriter struct {
}

func (w *stdOutWriter) Write(format string, a ...interface{}) error {
	_, err := fmt.Fprintf(os.Stdout, format, a...)
	return err
}

type stdErrWriter struct {
}

func (w *stdErrWriter) Write(format string, a ...interface{}) error {
	_, err := fmt.Fprintf(os.Stderr, format, a...)
	return err
}

type logWriter struct {
}

func (w *logWriter) Write(format string, a ...interface{}) error {
	config.Config().Logger().Infof(format, a...)
	return nil
}
