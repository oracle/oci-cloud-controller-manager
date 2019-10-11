// Copyright (c) 2017 Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"oracle.com/oci/httpsigner"
)

func TestNewTokenParser(t *testing.T) {
	ks := &mockKeySupplier{}
	as := httpsigner.Algorithms{}
	tp := NewTokenParser(ks, as)
	tpPtr := tp.(*tokenParser)
	assert.Equal(t, ks, tpPtr.keySupplier)
	assert.Equal(t, as, tpPtr.algSupplier)
}

func TestNewTokenParserBadValues(t *testing.T) {
	testIO := []struct {
		tc string
		ks httpsigner.KeySupplier
		as httpsigner.AlgorithmSupplier
	}{
		{tc: `should panic when key supplier is nil`,
			ks: nil, as: httpsigner.Algorithms{}},
		{tc: `should panic when algorithm supplier is nil`,
			ks: &mockKeySupplier{}, as: nil},
		{tc: `should panic when both key and algorith suppliers are nil`,
			ks: nil, as: nil},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			assert.Panics(t, func() { NewTokenParser(test.ks, test.as) })
		})
	}
}

// verify that tokenParser.Parse delegates to package function
func TestTokenParserParse(t *testing.T) {
	// setup
	ks := &mockKeySupplier{key: ""}
	alg := &mockAlgorithm{err: errTest}
	as := httpsigner.Algorithms{"RS256": alg}
	tp := NewTokenParser(ks, as)

	// go
	token, err := tp.Parse(testTokenStr)

	// validate
	assert.Nil(t, token)
	assert.NotNil(t, err)
	assert.Equal(t, errTest, err)
	assert.True(t, ks.KeyCalled)
	assert.True(t, alg.VerifyCalled)
}
