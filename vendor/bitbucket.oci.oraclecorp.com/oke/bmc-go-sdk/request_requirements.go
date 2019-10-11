// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

type identityCreationRequirement struct {
	CompartmentID string `header:"-" json:"compartmentId" url:"-"`
	Description   string `header:"-" json:"description" url:"-"`
	Name          string `header:"-" json:"name" url:"-"`
}

type ocidRequirement struct {
	CompartmentID string `header:"-" json:"compartmentId" url:"compartmentId"`
}

type listOCIDRequirement struct {
	CompartmentID string `header:"-" json:"-" url:"compartmentId"`
}

// Body is handled explicitly during marshal. Body can either be a []byte or io.ReadSeeker
type bodyMarshaller interface {
	body() interface{}
}

type bodyRequirement struct {
	Body interface{} `header:"-" json:"-" url:"-"`
}

func (b bodyRequirement) body() interface{} {
	return b.Body
}
