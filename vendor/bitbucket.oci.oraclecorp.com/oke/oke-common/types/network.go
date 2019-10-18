package types

import (
	baremetal "bitbucket.oci.oraclecorp.com/oke/bmc-go-sdk"
)

// VCNListRequestV1 is used for listing VCNs for a specified
// compartment
type VCNListRequestV1 struct {
	CompartmentOCID string `json:"compartmentOcid"`
}

// VCNListResponseV1 is the list response for vcns
type VCNListResponseV1 struct {
	VCNs []baremetal.VirtualNetwork `json:"vcns"`
}

// SubnetListRequestV1 requests subnets based on parameters
type SubnetListRequestV1 struct {
	CompartmentOCID string `json:"compartmentOcid"`
	VCNOCID         string `json:"vcnOcid"`
}

// SubnetListResponseV1 lists subnets for the response
type SubnetListResponseV1 struct {
	Subnets []baremetal.Subnet `json:"subnets"`
}

// ADListRequest retrieves ADs belonging to a tenancy
type ADListRequestV1 struct {
	CompartmentOCID string `json:"compartmentOcid"`
}

// ADListResponseV1 lists availability domains
type ADListResponseV1 struct {
	ADs []baremetal.AvailabilityDomain `json:"ads"`
}

// CompartListRequestV1 retrieves compartments based on tenancy
type CompartmentListRequestV1 struct {
}

// CompartmentListResponseV1 responds with compartments matching
// a tenancy
type CompartmentListResponseV1 struct {
	Compartments []baremetal.Compartment `json:"compartments"`
}
