// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"strconv"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ociRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oci_requests_total",
			Help: "OCI API requests total.",
		},
		[]string{"resource", "code", "verb"},
	)
)

type resource string

const (
	instanceResource       resource = "instance"
	vnicAttachmentResource resource = "vnic_attachment"
	vnicResource           resource = "vnic"
	subnetResource         resource = "subnet"
	loadBalancerResource   resource = "load_balancer"
	backendSetResource     resource = "load_balancer_backend_set"
	listenerResource       resource = "load_balancer_listener"
	certificateResource    resource = "load_balancer_certificate"
	workRequestResource    resource = "load_balancer_work_request"
	securityListResource   resource = "security_list"
	privateIPResource      resource = "ip_private"
	publicIPResource       resource = "ip_public"
	routeTableResource     resource = "route_table"
)

type verb string

const (
	getVerb    verb = "get"
	listVerb   verb = "list"
	createVerb verb = "create"
	updateVerb verb = "update"
	deleteVerb verb = "delete"
)

func incRequestCounter(err error, v verb, r resource) {
	statusCode := 200
	if err != nil {
		if serviceErr, ok := err.(common.ServiceError); ok {
			statusCode = serviceErr.GetHTTPStatusCode()
		} else {
			statusCode = 555 // ¯\_(ツ)_/¯
		}
	}

	ociRequestCounter.With(prometheus.Labels{
		"resource": string(r),
		"verb":     string(v),
		"code":     strconv.Itoa(statusCode),
	}).Inc()
}

func init() {
	prometheus.MustRegister(ociRequestCounter)
}
