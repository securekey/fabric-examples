/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package printer

import "fmt"

const (
	indentSize = 3
)

type printer struct {
	Formatter Formatter
}

// newPrinter returns a new Printer of the given OutputFormat and WriterType
func newPrinter(format OutputFormat, writerType WriterType) *printer {
	return &printer{Formatter: NewFormatter(format, writerType)}
}

// newPrinterWithOpts returns a new Printer of the given OutputFormat and WriterType
func newPrinterWithOpts(format OutputFormat, writerType WriterType, opts *FormatterOpts) *printer {
	return &printer{Formatter: NewFormatterWithOpts(format, writerType, opts)}
}

// Print prints a formatted string
func (p *printer) Print(frmt string, vars ...interface{}) {
	if p.Formatter == nil {
		fmt.Printf(frmt, vars)
		return
	}
	p.Formatter.Print(frmt, vars...)
}

// Field prints a field/value
func (p *printer) Field(Field string, value interface{}) {
	p.Formatter.Field(Field, value)
}

// Element starts a new named element. (Should be followed by ElementEnd.)
func (p *printer) Element(element string) {
	p.Formatter.Element(element)
}

// ElementEnd ends an element that was started with ElementStart.
func (p *printer) ElementEnd() {
	p.Formatter.ElementEnd()
}

// Array starts an array. (Should be followed by ArrayEnd.)
func (p *printer) Array(element string) {
	p.Formatter.Array(element)
}

// ArrayEnd ends an array that was started with Array.
func (p *printer) ArrayEnd() {
	p.Formatter.ArrayEnd()
}

// Item starts an array item at the given index. (A previous call to Array must have been made.)
func (p *printer) Item(element string, index interface{}) {
	p.Formatter.Item(element, index)
}

// ItemEnd ends an array item that was started with Item.
func (p *printer) ItemEnd() {
	p.Formatter.ItemEnd()
}

// ItemValue prints an array value at the given index. (A previous call to Array must have been made.)
func (p *printer) ItemValue(element string, index interface{}, value interface{}) {
	p.Formatter.ItemValue(element, index, value)
}

// Value prints a value.
func (p *printer) Value(value interface{}) {
	p.Formatter.Value(value)
}

// PrintHeader prints a header.
func (p *printer) PrintHeader() {
	p.Formatter.PrintHeader()
}

// PrintFooter prints a footer.
func (p *printer) PrintFooter() {
	p.Formatter.PrintFooter()
}
