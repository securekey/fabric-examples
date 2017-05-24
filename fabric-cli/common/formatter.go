/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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

type displayFormatter struct {
	indent int
}

func (p *displayFormatter) Print(frmt string, vars ...interface{}) {
	format := fmt.Sprintf("%s%s\n", p.prefix(), frmt)
	fmt.Printf(format, vars...)
}

func (p *displayFormatter) Field(field string, value interface{}) {
	if value != nil {
		fmt.Printf("%s%s: %v\n", p.prefix(), field, value)
	} else {
		fmt.Printf("%s%s:\n", p.prefix(), field)
	}
}

func (p *displayFormatter) Element(element string) {
	if element != "" {
		fmt.Printf("%s%s:\n", p.prefix(), element)
	}
	p.indent++
}

func (p *displayFormatter) ElementEnd() {
	p.indent--
}

func (p *displayFormatter) Array(element string) {
	if element != "" {
		fmt.Printf("%s%s:\n", p.prefix(), element)
	}
	p.indent++
}

func (p *displayFormatter) ArrayEnd() {
	p.indent--

}

func (p *displayFormatter) Item(element string, index interface{}) {
	if element != "" {
		fmt.Printf("%s%s[%v]:\n", p.prefix(), element, index)
	}
	p.indent++
}

func (p *displayFormatter) ItemEnd() {
	p.ElementEnd()
}

func (p *displayFormatter) ItemValue(element string, index interface{}, value interface{}) {
	if element != "" {
		fmt.Printf("%s%s[%v]: %v\n", p.prefix(), element, index, value)
	}
}

func (p *displayFormatter) Value(value interface{}) {
	fmt.Printf("%s%v", p.prefix(), value)
}

func (p *displayFormatter) PrintHeader() {
	fmt.Printf("%s\n", strings.Repeat("*", 100))
}

func (p *displayFormatter) PrintFooter() {
	fmt.Printf("%s\n", strings.Repeat("*", 100))
}

func (p *displayFormatter) prefix() string {
	s := "***** "
	for i := 0; i < p.indent; i++ {
		s = s + fmt.Sprintf("|%s", strings.Repeat(" ", indentSize-1))
	}
	return s
}

type jsonFormatter struct {
	commaRequired bool
}

func (p *jsonFormatter) Print(frmt string, vars ...interface{}) {
	// Ignore additional output in JSON format
}

func (p *jsonFormatter) Field(field string, value interface{}) {
	if p.commaRequired {
		fmt.Printf(",")
	}

	if value != nil {
		fmt.Printf("\"%s\":\"%v\"", field, value)
	} else {
		fmt.Printf("\"%s\":\"null\"", field)
	}

	p.commaRequired = true
}

func (p *jsonFormatter) Element(element string) {
	if p.commaRequired {
		fmt.Printf(",")
		p.commaRequired = false
	}
	if element != "" {
		fmt.Printf("\"%s\":{", element)
	} else {
		fmt.Printf("{")
	}
}

func (p *jsonFormatter) ElementEnd() {
	fmt.Printf("}")
	p.commaRequired = true
}

func (p *jsonFormatter) Array(element string) {
	if p.commaRequired {
		fmt.Printf(",")
		p.commaRequired = false
	}
	if element != "" {
		fmt.Printf("\"%s\":[", element)
	} else {
		fmt.Printf("[")
	}
}

func (p *jsonFormatter) ArrayEnd() {
	fmt.Printf("]")
	p.commaRequired = true
}

func (p *jsonFormatter) Item(element string, index interface{}) {
	if p.commaRequired {
		fmt.Printf(",")
		p.commaRequired = false
	}
	fmt.Printf("{")
}

func (p *jsonFormatter) ItemEnd() {
	p.ElementEnd()
}

func (p *jsonFormatter) ItemValue(element string, index interface{}, value interface{}) {
	if p.commaRequired {
		fmt.Printf(",")
	}
	fmt.Printf("\"%v\"", value)
}

func (p *jsonFormatter) Value(value interface{}) {
	if p.commaRequired {
		fmt.Printf(",")
	}
	fmt.Printf("\"%v\"", value)
}

func (p *jsonFormatter) PrintHeader() {
	p.Element("")
	p.commaRequired = false
}

func (p *jsonFormatter) PrintFooter() {
	p.ElementEnd()
	fmt.Printf("\n")
	p.commaRequired = false
}
