// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rsa"

	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMemorySessionKeySupplierPrivateKey ensures that PrivateKey() returns an *rsa.PrivateKey
func TestMemorySessionKeySupplierPrivateKey(t *testing.T) {
	k := NewMemorySessionKeySupplier()
	p := k.PrivateKey()

	var keyType *rsa.PrivateKey
	assert.IsType(t, keyType, p)
}

// TestMemorySessionKeySupplierPublicKey ensures that PublicKey() returns an *rsa.PublicKey
func TestMemorySessionKeySupplierPublicKey(t *testing.T) {
	k := NewMemorySessionKeySupplier()
	p := k.PublicKey()

	var keyType *rsa.PublicKey
	assert.IsType(t, keyType, p)
}

// TestMemorySessionKeySupplierGeneratesKeyOnDemand ensures that a key pair is generated
// on demand when PrivateKey() is called
func TestMemorySessionKeySupplierGeneratesKeyOnDemand(t *testing.T) {
	k := NewMemorySessionKeySupplier()
	assert.Nil(t, k.privateKey)
	k.PrivateKey()
	assert.NotNil(t, k.privateKey)
}
