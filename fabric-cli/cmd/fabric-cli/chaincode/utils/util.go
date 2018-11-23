/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"fmt"
	"github.com/pkg/errors"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	randFunc = "$rand("
	padFunc  = "$pad("
)

// AsBytes converts the string array to an array of byte arrays.
// The args may contain functions $rand(n) or $pad(n,chars).
// The functions are evaluated before returning.
//
// Examples:
// - "key$rand(3)" -> "key0" or "key1" or "key2"
// - "val$pad(3,XYZ) -> "valXYZXYZXYZ"
// - "val$pad($rand(3),XYZ) -> "val" or "valXYZ or "valXYZXYZ"
func AsBytes(args []string) [][]byte {
	rand := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	bytes := make([][]byte, len(args))

	if cliconfig.Config().Verbose() {
		fmt.Printf("Args:\n")
	}
	for i, a := range args {
		arg := getArg(rand, a)
		if cliconfig.Config().Verbose() {
			fmt.Printf("- [%d]=%s\n", i, arg)
		}
		bytes[i] = []byte(arg)
	}
	return bytes
}

func getArg(r *rand.Rand, arg string) string {
	arg = evaluateRandExpression(r, arg)
	arg = evaluatePadExpression(arg)
	return arg
}

// evaluateRandExpression replaces occurrences of $rand(n) with a random
// number between 0 and n (exclusive)
func evaluateRandExpression(r *rand.Rand, arg string) string {
	return evaluateExpression(arg, randFunc,
		func(expression string) (string, error) {
			n, err := strconv.Atoi(expression)
			if err != nil {
				return "", errors.Errorf("invalid number %s in $rand expression\n", expression)
			}
			return strconv.Itoa(r.Intn(n)), nil
		})
}

// evaluatePadExpression replaces occurrences of $pad(n,chars) with n of the given pad characters
func evaluatePadExpression(arg string) string {
	return evaluateExpression(arg, padFunc,
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

func evaluateExpression(expression, funcType string, evaluate func(string) (string, error)) string {
	result := ""
	for {
		i := strings.Index(expression, funcType)
		if i == -1 {
			result += expression
			break
		}

		j := strings.Index(expression[i:], ")")
		if j == -1 {
			fmt.Printf("expecting ')' in expression '%s'", expression)
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
