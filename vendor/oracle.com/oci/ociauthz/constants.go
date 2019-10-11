// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"time"
)

const (
	// Default timeout to use in HTTP calls to the identity service
	defaultTimeout = 10 * time.Second

	// How long in minutes should the key service cache customer public keys for?
	keyServiceDefaultCachePeriodMinutes = 2

	// How long in minutes should the key service cache token service public keys for?
	keyServiceTokenServiceCachePeriodMinutes = 180

	// RegionalADValue is the default value of PhysicalAD and should be used when the AD is regional
	RegionalADValue = `all`

	// STSTokenPrefix is the prefix (ST$) used for STS tokens
	STSTokenPrefix = `ST$`

	// KeyIDForceRotate is a keyID used to force rotation of the current key during key lookup
	KeyIDForceRotate = "KEY_ID_FORCE_ROTATE"
)

// Identity service paths
const (

	// Path to obtain an STS token
	x509URITemplate = `%s/x509`

	// Path to lookup STS Service public key
	keyServiceURITemplate = `%s/keys`

	// Path to lookup customer API public keys
	keyServiceSRURITemplate = `%s/SR/keys`

	// Path to make thin-client authorization requests to
	authorizeURITemplate = `%s/authorization/authorizerequest`

	// Path to make thin-client authorization with a tag slug
	authorizeWithTagsURITemplate = `%s/authorization/authorizerequest2`

	// Path to make thin-client associationAuthorization requests to
	associationAuthorizeURITemplate = `%s/authorization/associaterequest`

	// Path to obtain an OBO Token
	OboURITemplate = `%s/obo`
)
