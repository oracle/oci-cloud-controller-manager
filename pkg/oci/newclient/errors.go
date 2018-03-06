package client

import (
	"net/http"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/pkg/errors"
)

var (
	errNotFound = errors.New("not found")
)

// IsNotFound returns true if the given error indicates that a resource could
// not be found.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	err = errors.Cause(err)
	if err == errNotFound {
		return true
	}

	serviceErr, ok := common.IsServiceError(err)
	return ok && serviceErr.GetHTTPStatusCode() == http.StatusNotFound
}
