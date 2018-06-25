/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

const (
	randomVar = "$rand("
)

// AsBytes converts the string array to an array of byte arrays.
// If the arg contains the string $rand(n) then $rand(n) is replaced
// with a random number between 0 and n-1.
func AsBytes(args []string) [][]byte {
	bytes := make([][]byte, len(args))
	for i, arg := range args {
		bytes[i] = []byte(getArg(arg))
	}
	return bytes
}

func getArg(arg string) string {
	if !strings.Contains(arg, randomVar) {
		return arg
	}

	i := strings.Index(arg, randomVar)
	j := strings.Index(arg, ")")
	ns := arg[i+len(randomVar) : j]
	n, err := strconv.Atoi(ns)
	if err != nil {
		fmt.Printf("Error converting random value [%s] to int: %s\n", ns, err)
		return arg
	}

	return arg[0:i] + strconv.Itoa(rand.Intn(n)) + arg[j:len(arg)-1]
}
