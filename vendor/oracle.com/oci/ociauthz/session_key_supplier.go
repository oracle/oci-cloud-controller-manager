// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rand"
	"crypto/rsa"
)

// SessionKeySupplier generates and caches a key pair
type SessionKeySupplier interface {
	PublicKey() *rsa.PublicKey
	PrivateKey() *rsa.PrivateKey
	RefreshKeys()
}

// MemorySessionKeySupplier implements SessionKeySupplier and provides in-memory
// caching of the key pair
type MemorySessionKeySupplier struct {
	privateKey *rsa.PrivateKey
}

// NewMemorySessionKeySupplier returns a new instance of the memory session key supplier
func NewMemorySessionKeySupplier() *MemorySessionKeySupplier {
	return &MemorySessionKeySupplier{}
}

// RefreshKeys regenerates the key pair
func (k *MemorySessionKeySupplier) RefreshKeys() {
	// TODO clear k.privateKey using ClearSecret()

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	k.privateKey = privateKey
}

// PrivateKey returns the private key from the key pair
func (k *MemorySessionKeySupplier) PrivateKey() *rsa.PrivateKey {
	if k.privateKey == nil {
		k.RefreshKeys()
	}

	return k.privateKey
}

// PublicKey returns the public key from the key pair
func (k *MemorySessionKeySupplier) PublicKey() *rsa.PublicKey {
	if k.privateKey == nil {
		k.RefreshKeys()
	}

	return &k.privateKey.PublicKey
}
