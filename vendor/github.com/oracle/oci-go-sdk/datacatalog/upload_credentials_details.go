// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DataCatalog API
//
// A description of the DataCatalog API
//

package datacatalog

import (
	"github.com/oracle/oci-go-sdk/common"
)

// UploadCredentialsDetails Upload credential file and connection metadata
type UploadCredentialsDetails struct {

	// Information used in updating connection credentials
	CredentialPayload []byte `mandatory:"true" json:"credentialPayload"`

	ConnectionDetail *UpdateConnectionDetails `mandatory:"false" json:"connectionDetail"`
}

func (m UploadCredentialsDetails) String() string {
	return common.PointerString(m)
}
