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
