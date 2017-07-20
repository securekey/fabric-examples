/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"fmt"
	"strings"
)

// OutputFormat specifies the format for printing data
type OutputFormat uint8

const (
	// RAW displays the raw data
	RAW OutputFormat = iota

	// JSON formats the data into Java Script Object Notation format
	JSON

	// DISPLAY formats the data into a human readable format
	DISPLAY
)

func (f OutputFormat) String() string {
	switch f {
	case DISPLAY:
		return "display"
	case JSON:
		return "json"
	case RAW:
		return "raw"
	default:
		return "unknown"
	}
}

// AsOutputFormat returns the OutputFormat given an Output Format string
func AsOutputFormat(f string) OutputFormat {
	switch strings.ToLower(f) {
	case "json":
		return JSON
	case "raw":
		return RAW
	default:
		return DISPLAY
	}
}

// Formatter outputs the data in a specific format
type Formatter interface {
	// PrintHeader outputs a header
	PrintHeader()

	// PrintFooter outputs a footer
	PrintFooter()

	// Element starts an element
	Element(element string)

	// ElementEnd ends an element
	ElementEnd()

	// Field outputs the value of a field
	Field(field string, value interface{})

	// Array starts an array
	Array(element string)

	// ArrayEnd ends the array
	ArrayEnd()

	// Item starts a complex array item
	Item(element string, index interface{})

	// ItemEnd ends an array item
	ItemEnd()

	// ItemValue outputs a simple array item
	ItemValue(element string, index interface{}, value interface{})

	// Value outputs a value
	Value(value interface{})

	// Print outputs additional info
	Print(frmt string, vars ...interface{})
}

// NewFormatter returns a new Formatter given the format and writer type. nil is returned
// if no formatter exists for the given type
func NewFormatter(format OutputFormat, writerType WriterType) Formatter {
	switch format {
	case JSON:
		return &jsonFormatter{formatter: formatter{writer: NewWriter(writerType)}}
	case DISPLAY:
		return &displayFormatter{formatter: formatter{writer: NewWriter(writerType)}}
	default:
		return nil
	}
}

type formatter struct {
	writer Writer
}

func (f *formatter) write(format string, a ...interface{}) error {
	return f.writer.Write(format, a...)
}

type displayFormatter struct {
	formatter
	indent int
}

func (p *displayFormatter) Print(frmt string, vars ...interface{}) {
	format := fmt.Sprintf("%s%s\n", p.prefix(), frmt)
	p.write(format, vars...)
}

func (p *displayFormatter) Field(field string, value interface{}) {
	if value != nil {
		p.write("%s%s: %v\n", p.prefix(), field, value)
	} else {
		p.write("%s%s:\n", p.prefix(), field)
	}
}

func (p *displayFormatter) Element(element string) {
	if element != "" {
		p.write("%s%s:\n", p.prefix(), element)
	}
	p.indent++
}

func (p *displayFormatter) ElementEnd() {
	p.indent--
}

func (p *displayFormatter) Array(element string) {
	if element != "" {
		p.write("%s%s:\n", p.prefix(), element)
	}
	p.indent++
}

func (p *displayFormatter) ArrayEnd() {
	p.indent--

}

func (p *displayFormatter) Item(element string, index interface{}) {
	if element != "" {
		p.write("%s%s[%v]:\n", p.prefix(), element, index)
	}
	p.indent++
}

func (p *displayFormatter) ItemEnd() {
	p.ElementEnd()
}

func (p *displayFormatter) ItemValue(element string, index interface{}, value interface{}) {
	if element != "" {
		p.write("%s%s[%v]: %v\n", p.prefix(), element, index, value)
	}
}

func (p *displayFormatter) Value(value interface{}) {
	p.write("%s%v", p.prefix(), value)
}

func (p *displayFormatter) PrintHeader() {
	p.write("%s\n", strings.Repeat("*", 100))
}

func (p *displayFormatter) PrintFooter() {
	p.write("%s\n", strings.Repeat("*", 100))
}

func (p *displayFormatter) prefix() string {
	s := "***** "
	for i := 0; i < p.indent; i++ {
		s = s + fmt.Sprintf("|%s", strings.Repeat(" ", indentSize-1))
	}
	return s
}

type jsonFormatter struct {
	formatter
	commaRequired bool
}

func (p *jsonFormatter) Print(frmt string, vars ...interface{}) {
	// Ignore additional output in JSON format
}

func (p *jsonFormatter) Field(field string, value interface{}) {
	if p.commaRequired {
		p.write(",")
	}

	if value != nil {
		p.write("\"%s\":\"%v\"", field, value)
	} else {
		p.write("\"%s\":\"null\"", field)
	}

	p.commaRequired = true
}

func (p *jsonFormatter) Element(element string) {
	if p.commaRequired {
		p.write(",")
		p.commaRequired = false
	}
	if element != "" {
		p.write("\"%s\":{", element)
	} else {
		p.write("{")
	}
}

func (p *jsonFormatter) ElementEnd() {
	p.write("}")
	p.commaRequired = true
}

func (p *jsonFormatter) Array(element string) {
	if p.commaRequired {
		p.write(",")
		p.commaRequired = false
	}
	if element != "" {
		p.write("\"%s\":[", element)
	} else {
		p.write("[")
	}
}

func (p *jsonFormatter) ArrayEnd() {
	p.write("]")
	p.commaRequired = true
}

func (p *jsonFormatter) Item(element string, index interface{}) {
	if p.commaRequired {
		p.write(",")
		p.commaRequired = false
	}
	p.write("{")
}

func (p *jsonFormatter) ItemEnd() {
	p.ElementEnd()
}

func (p *jsonFormatter) ItemValue(element string, index interface{}, value interface{}) {
	if p.commaRequired {
		p.write(",")
	}
	p.write("\"%v\"", value)
}

func (p *jsonFormatter) Value(value interface{}) {
	if p.commaRequired {
		p.write(",")
	}
	p.write("\"%v\"", value)
}

func (p *jsonFormatter) PrintHeader() {
	p.Element("")
	p.commaRequired = false
}

func (p *jsonFormatter) PrintFooter() {
	p.ElementEnd()
	p.write("\n")
	p.commaRequired = false
}
