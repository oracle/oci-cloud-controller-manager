// Copyright (c) 2017 Oracle and/or its affiliates. All rights reserved.

package httpsigner

// JWT algorithm names (from: https://tools.ietf.org/html/draft-jones-json-web-signature-04#section-6)
const (
	JWTRSASHA256 = "RS256"
)

// JWTAlgorithms is an AlgorithmSupplier indexed by JWT standard algorithm names
var JWTAlgorithms = Algorithms{
	JWTRSASHA256: AlgorithmRSASHA256,
}
