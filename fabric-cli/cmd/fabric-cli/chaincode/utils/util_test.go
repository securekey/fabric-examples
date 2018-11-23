/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
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

func TestGetArg(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	result := getArg(rand, "Value_$pad(3,X)!")
	assert.Equal(t, "Value_XXX!", result)

	for i := 0; i < 5; i++ {
		result := getArg(rand, "Value_$rand(2)!")
		assert.True(t, result == "Value_0!" || result == "Value_1!")
	}

	for i := 0; i < 5; i++ {
		result := getArg(rand, "Value_$pad(3,X)_$rand(2)!")
		assert.True(t, result == "Value_XXX_0!" || result == "Value_XXX_1!")
	}

	for i := 0; i < 5; i++ {
		result := getArg(rand, "Value_$pad($rand(3),X)!")
		assert.True(t, result == "Value_!" || result == "Value_X!" || result == "Value_XX!")
	}
}
