package util

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	metricErrors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// CompartmentIDAnnotation is the annotation for OCI compartment
	CompartmentIDAnnotation = "oci.oraclecloud.com/compartment-id"

	// Error codes
	Err429           = "429"
	Err4XX           = "4XX"
	Err5XX           = "5XX"
	ErrValidation    = "VALIDATION_ERROR"
	ErrLimitExceeded = "LIMIT_EXCEEDED"
	Success          = "SUCCESS"

	// Components generating errors
	// Load Balancer
	LoadBalancerType = "LB"
	// storage types
	CSIStorageType = "CSI"
	FVDStorageType = "FVD"
)

// LookupNodeCompartment returns the compartment OCID for the given nodeName.
func LookupNodeCompartment(k kubernetes.Interface, nodeName string) (string, error) {
	node, err := k.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if compartmentID, ok := node.ObjectMeta.Annotations[CompartmentIDAnnotation]; ok {
		if compartmentID != "" {
			return compartmentID, nil
		}
	}
	return "", errors.New("CompartmentID annotation is not present")
}

func GetError(err error) string {
	if err == nil {
		return ""
	}
	err = metricErrors.Cause(err)

	cause := err.Error()
	if cause == "" {
		return ""
	}

	re := regexp.MustCompile(`http status code:\s*(\d+)`)
	if match := re.FindStringSubmatch(cause); match != nil {
		if status, er := strconv.Atoi(match[1]); er == nil {
			if status >= 500 {
				return Err5XX
			} else if status >= 400 {
				if strings.Contains(err.Error(), "Service error:LimitExceeded") {
					return ErrLimitExceeded
				}
				if status == 429 {
					return Err429
				}
				return Err4XX
			}
		}
	}
	return ErrValidation
}

func GetMetricDimensionForComponent(err string, component string) string {
	if err == "" || component == "" {
		return ""
	}
	return fmt.Sprintf("%s_%s", component, err)
}
