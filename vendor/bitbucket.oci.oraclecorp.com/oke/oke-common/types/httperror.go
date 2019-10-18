package types

import "fmt"

// HTTPError holds details about an API error
type HTTPError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status"`
}

// Error serializes the HTTPError
func (e *HTTPError) Error() string {
	return fmt.Sprintf("http status [%d]: %s", e.StatusCode, e.Message)
}
