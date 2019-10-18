// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrorInvalidArgErrorString verifies the ErrorInvalidArg Error function produces the expected string
func TestErrorInvalidArgErrorString(t *testing.T) {
	argName := "arg1"
	e := newErrorInvalidArg(argName)
	assert.Equal(t, "httpsigner: Invalid argument value for 'arg1'", e.Error())
}

// TestKeyRotationError verifies replacement keyid can be placed inside the error object
func TestKeyRotationError(t *testing.T) {
	replacementID := "ReplacementKeyID"
	oldID := "OldKeyID"
	var e error = NewKeyRotationError(replacementID, oldID)
	assert.Equal(t, "Requested key has expired and has been automatically rotated.", e.Error())
	assert.Equal(t, replacementID, e.(*KeyRotationError).ReplacementKeyID)
	assert.Equal(t, oldID, e.(*KeyRotationError).OldKeyID)
}
