package types

import "fmt"

// RequestID represents the `opc-request-id` HTTP headers
type RequestID struct {
	CustomerRequestID   string
	CallStackID         string
	IndividualRequestID string
}

func (r *RequestID) String() string {
	s := r.CustomerRequestID
	if len(r.CallStackID) > 0 {
		s = fmt.Sprintf("%s/%s", s, r.CallStackID)
	}
	if len(r.IndividualRequestID) > 0 {
		s = fmt.Sprintf("%s/%s", s, r.IndividualRequestID)
	}
	return s
}

// DowstreamString generates a string that can be used as a
// downstream request ID. Returns an error if there is not
// enough information in the request ID to generate the
// downstream request ID.
func (r *RequestID) DowstreamString() (string, error) {
	if len(r.CallStackID) <= 0 {
		return "", fmt.Errorf("the CallStackID must be set in order to create a downstream string")
	}
	s := fmt.Sprintf("%s/%s", r.CustomerRequestID, r.CallStackID)
	return s, nil
}
