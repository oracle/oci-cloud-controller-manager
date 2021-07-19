package util

import (
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CompartmentIDAnnotation is the annotation for OCI compartment
const CompartmentIDAnnotation = "oci.oraclecloud.com/compartment-id"

// LookupNodeCompartment returns the compartment OCID for the given nodeName.
func LookupNodeCompartment(k kubernetes.Interface, nodeName string) (string, error) {
	node, err := k.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
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
