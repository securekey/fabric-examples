/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
)

const (
	randFunc = "$rand("
	padFunc  = "$pad("
	seqFunc  = "$seq("
	setFunc  = "$set("
	varExp   = "${"
)

var (
	sequence uint64
)

type Context interface {
	SetVar(name, value string)
	GetVar(name string) (string, bool)
}

// AsBytes converts the string array to an array of byte arrays.
// The args may contain functions $rand(n) or $pad(n,chars).
// The functions are evaluated before returning.
//
// Examples:
// - "key$rand(3)" -> "key0" or "key1" or "key2"
// - "val$pad(3,XYZ) -> "valXYZXYZXYZ"
// - "val$pad($rand(3),XYZ) -> "val" or "valXYZ" or "valXYZXYZ"
// - "key$seq()" -> "key1", "key2", "key2", ...
// - "val$pad($seq(),X)" -> "valX", "valXX", "valXX", "valXXX", ...
// - "Key_$set(x,$seq())=Val_${x}" -> Key_1=Val_1, Key_2=Val_2, ...
func AsBytes(ctxt Context, args []string) [][]byte {
	rand := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	bytes := make([][]byte, len(args))

	if cliconfig.Config().Verbose() {
		fmt.Printf("Args:\n")
	}
	for i, a := range args {
		arg := getArg(ctxt, rand, a)
		if cliconfig.Config().Verbose() {
			fmt.Printf("- [%d]=%s\n", i, arg)
		}
		bytes[i] = []byte(arg)
	}
	return bytes
}

func getArg(ctxt Context, r *rand.Rand, arg string) string {
	arg = evaluateSeqExpression(arg)
	arg = evaluateRandExpression(r, arg)
	arg = evaluatePadExpression(arg)
	arg = evaluateSetExpression(ctxt, arg)
	arg = evaluateVarExpression(ctxt, arg)
	return arg
}

// evaluateSeqExpression replaces occurrences of $seq() with a sequential
// number starting at 1 and incrementing for each task
func evaluateSeqExpression(arg string) string {
	return evaluateExpression(arg, seqFunc, ")",
		func(expression string) (string, error) {
			return strconv.FormatUint(atomic.AddUint64(&sequence, 1), 10), nil
		})
}

// evaluateRandExpression replaces occurrences of $rand(n) with a random
// number between 0 and n (exclusive)
func evaluateRandExpression(r *rand.Rand, arg string) string {
	return evaluateExpression(arg, randFunc, ")",
		func(expression string) (string, error) {
			n, err := strconv.ParseInt(expression, 10, 64)
			if err != nil {
				return "", errors.Errorf("invalid number %s in $rand expression\n", expression)
			}
			return strconv.FormatInt(r.Int63n(n), 10), nil
		})
}

// evaluatePadExpression replaces occurrences of $pad(n,chars) with n of the given pad characters
func evaluatePadExpression(arg string) string {
	return evaluateExpression(arg, padFunc, ")",
		func(expression string) (string, error) {
			s := strings.Split(expression, ",")
			if len(s) != 2 {
				return "", errors.Errorf("invalid $pad expression: '%s'. Expecting $pad(n,chars)", expression)
			}

			n, err := strconv.Atoi(s[0])
			if err != nil {
				return "", errors.Errorf("invalid number %s in $pad expression\n", s[0])
			}

			result := ""
			for i := 0; i < n; i++ {
				result += s[1]
			}

			return result, nil
		})
}

// evaluateSetExpression sets a variable to the given value using the syntax $set(var,expression).
// The variable may be used in a subsequent expression, ${var}.
// Example: $set(v,$rand(10)) sets variable "v" to a random value that may be accessed as ${v}
func evaluateSetExpression(ctxt Context, arg string) string {
	return evaluateExpression(arg, setFunc, ")",
		func(expression string) (string, error) {
			s := strings.Split(expression, ",")
			if len(s) != 2 {
				return "", errors.Errorf("invalid $set expression: '%s'. Expecting $set(var,value)", expression)
			}

			variable := s[0]
			value := s[1]

			ctxt.SetVar(variable, value)
			return value, nil
		})
}

// evaluateVarExpression references a variable previously set with the $set(var,expression) expression.
// Example: ${someVar}
func evaluateVarExpression(ctxt Context, arg string) string {
	return evaluateExpression(arg, varExp, "}",
		func(varName string) (string, error) {
			value, ok := ctxt.GetVar(varName)
			if !ok {
				return "", errors.Errorf("variable [%s] not set", varName)
			}
			return value, nil
		})
}

func evaluateExpression(expression, funcType, endDelim string, evaluate func(string) (string, error)) string {
	result := ""
	for {
		i := strings.Index(expression, funcType)
		if i == -1 {
			result += expression
			break
		}

		j := strings.Index(expression[i:], endDelim)
		if j == -1 {
			fmt.Printf("expecting '%s' in expression '%s'", endDelim, expression)
			result = expression
			break
		}

		j = i + j

		replacement, err := evaluate(expression[i+len(funcType) : j])
		if err != nil {
			fmt.Printf("%s\n", err)
			result += expression[0 : j+1]
		} else {
			result += expression[0:i] + replacement
		}

		expression = expression[j+1:]
	}

	return result
}

func NewContext() Context {
	return &defaultContext{
		vars: make(map[string]string),
	}
}

type defaultContext struct {
	vars map[string]string
}

func (c *defaultContext) SetVar(k, v string) {
	c.vars[k] = v
}

func (c *defaultContext) GetVar(k string) (string, bool) {
	value, ok := c.vars[k]
	return value, ok
}
