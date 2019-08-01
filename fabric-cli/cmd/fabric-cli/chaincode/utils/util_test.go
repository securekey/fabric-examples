/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestEvaluatePadExpression(t *testing.T) {
	result := evaluatePadExpression("Value_$pad(3,XYZ)!")
	assert.Equal(t, "Value_XYZXYZXYZ!", result)

	result = evaluatePadExpression("Value_$pad(3,XYZ)_$pad(2,123)!")
	assert.Equal(t, "Value_XYZXYZXYZ_123123!", result)
}

func TestFailEvaluatePadExpression(t *testing.T) {
	result := evaluatePadExpression("Value_$pad(3,XYZ!")
	assert.Equal(t, "Value_$pad(3,XYZ!", result)

	result = evaluatePadExpression("Value_$pad3,XYZ)!")
	assert.Equal(t, "Value_$pad3,XYZ)!", result)

	result = evaluatePadExpression("Value_$pad(3)!")
	assert.Equal(t, "Value_$pad(3)!", result)

	result = evaluatePadExpression("Value_$pad(3)_$pad(3,X)!")
	assert.Equal(t, "Value_$pad(3)_XXX!", result)
}

func TestEvaluateRandExpression(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	for i := 0; i < 5; i++ {
		result := evaluateRandExpression(rand, "Value_$rand(2)!")
		assert.True(t, result == "Value_0!" || result == "Value_1!")

		result = evaluateRandExpression(rand, "Value_$rand(2)_$rand(1)!")
		assert.True(t, result == "Value_0_0!" || result == "Value_1_0!")
	}
}

func TestFailEvaluateRandExpression(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	result := evaluateRandExpression(rand, "Value_$rand(3,!")
	assert.Equal(t, "Value_$rand(3,!", result)

	result = evaluateRandExpression(rand, "Value_$rand3)!")
	assert.Equal(t, "Value_$rand3)!", result)

	result = evaluateRandExpression(rand, "Value_$rand(X)!")
	assert.Equal(t, "Value_$rand(X)!", result)

	result = evaluateRandExpression(rand, "Value_$rand(X)_$rand(1)!")
	assert.Equal(t, "Value_$rand(X)_0!", result)
}

func TestEvaluateSeqExpression(t *testing.T) {
	assert.Equal(t, "Value_1!", evaluateSeqExpression("Value_$seq()!"))
	assert.Equal(t, "Value_2!", evaluateSeqExpression("Value_$seq()!"))
	assert.Equal(t, "Value_3!", evaluateSeqExpression("Value_$seq()!"))
}

func TestFailEvaluateSeqExpression(t *testing.T) {
	assert.Equal(t, "Value_$seq(!", evaluateSeqExpression("Value_$seq(!"))
}

func TestEvaluateFileExpression(t *testing.T) {
	assert.Equal(t, `{"Field1": "Value1"}`, evaluateFileExpression("$file(./test.json)"))
}

func TestFailEvaluateFileExpression(t *testing.T) {
	assert.Equal(t, "$file(./invalid.json)", evaluateSeqExpression("$file(./invalid.json)"))
}

func TestEvaluateSetExpression(t *testing.T) {
	ctxt := NewContext()
	assert.Equal(t, "Value_1000!", evaluateSetExpression(ctxt, "Value_$set(x,1000)!"))
	assert.Equal(t, "Value_1000!", evaluateVarExpression(ctxt, "Value_${x}!"))
}

func TestFailEvaluateSetExpression(t *testing.T) {
	ctxt := NewContext()
	assert.Equal(t, "Value_$set(x,1000", evaluateSetExpression(ctxt, "Value_$set(x,1000"))
	assert.Equal(t, "Value_${x", evaluateVarExpression(ctxt, "Value_${x"))

	// Var not set
	assert.Equal(t, "Value_${x}", evaluateVarExpression(NewContext(), "Value_${x}"))
}

func TestGetArg(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	result := getArg(NewContext(), rand, "Value_$pad(3,X)!")
	assert.Equal(t, "Value_XXX!", result)

	for i := 0; i < 5; i++ {
		result := getArg(NewContext(), rand, "Value_$rand(2)!")
		assert.True(t, result == "Value_0!" || result == "Value_1!")
	}

	for i := 0; i < 5; i++ {
		result := getArg(NewContext(), rand, "Value_$pad(3,X)_$rand(2)!")
		assert.True(t, result == "Value_XXX_0!" || result == "Value_XXX_1!")
	}

	for i := 0; i < 5; i++ {
		result := getArg(NewContext(), rand, "Value_$pad($rand(3),X)!")
		assert.True(t, result == "Value_!" || result == "Value_X!" || result == "Value_XX!")
	}

	n := sequence + 1
	for i := n; i <= n+5; i++ {
		result := getArg(NewContext(), rand, "Value_$seq()!")
		assert.True(t, result == fmt.Sprintf("Value_%d!", i))
	}

	n = sequence + 1
	for i := n; i <= n+5; i++ {
		ctxt := NewContext()
		assert.Equal(t, fmt.Sprintf("Key_%d=Val_%d", i, i), getArg(ctxt, rand, "Key_$set(x,$seq())=Val_${x}"))
		assert.Equal(t, fmt.Sprintf("Value_%d!", i), getArg(ctxt, rand, "Value_${x}!"))
	}
}
