package apierrors

import (
	"fmt"
)

// ErrorV3 contains the details about an error, including standardized code
type ErrorV3 struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	Status       string `json:"status,omitempty"`
	OPCRequestID string `json:"opc-request-id,omitempty"`
}

// NewErrorV3 is a convenience function to create an ErrorV3
func NewErrorV3(code, message string) *ErrorV3 {
	return &ErrorV3{
		Code:    code,
		Message: message,
	}
}

// String returns the ErrorV3 as a string
func (e ErrorV3) String() string {
	return fmt.Sprintf("Code: %s; Message: %s", e.Code, e.Message)
}

// Error is provided to conform to the error interface
func (e ErrorV3) Error() string {
	return fmt.Sprintf("Code: %s; Message: %s", e.Code, e.Message)
}
