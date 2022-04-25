package client

import (
	"net/http"
	"testing"

	"github.com/oracle/oci-go-sdk/v50/common"
)

type mockServiceError struct {
	StatusCode   int
	Code         string
	Message      string
	OpcRequestID string
}

func (m mockServiceError) GetHTTPStatusCode() int {
	return m.StatusCode
}

func (m mockServiceError) GetMessage() string {
	return m.Message
}

func (m mockServiceError) GetCode() string {
	return m.Code
}

func (m mockServiceError) GetOpcRequestID() string {
	return m.OpcRequestID
}

func TestIsRetryableServiceError(t *testing.T) {
	testCases := map[string]struct {
		error    common.ServiceError
		expected bool
	}{
		"HTTP400RelatedResourceNotAuthorizedOrNotFound": {
			error: mockServiceError{
				StatusCode: http.StatusBadRequest,
				Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
			},
			expected: true,
		},
		"HTTP401NotAuthenticated": {
			error: mockServiceError{
				StatusCode: http.StatusUnauthorized,
				Code:       HTTP401NotAuthenticatedCode,
			},
			expected: true,
		},
		"HTTP404NotAuthorizedOrNotFound": {
			error: mockServiceError{
				StatusCode: http.StatusNotFound,
				Code:       HTTP404NotAuthorizedOrNotFoundCode,
			},
			expected: true,
		},
		"HTTP409IncorrectState": {
			error: mockServiceError{
				StatusCode: http.StatusConflict,
				Code:       HTTP409IncorrectStateCode,
			},
			expected: true,
		},
		"HTTP409NotAuthorizedOrResourceAlreadyExists": {
			error: mockServiceError{
				StatusCode: http.StatusConflict,
				Code:       HTTP409NotAuthorizedOrResourceAlreadyExistsCode,
			},
			expected: true,
		},
		"HTTP429TooManyRequests": {
			error: mockServiceError{
				StatusCode: http.StatusTooManyRequests,
				Code:       HTTP429TooManyRequestsCode,
			},
			expected: true,
		},
		"HTTP500InternalServerError": {
			error: mockServiceError{
				StatusCode: http.StatusInternalServerError,
				Code:       HTTP500InternalServerErrorCode,
			},
			expected: true,
		},
		"NonRetryable": {
			error: mockServiceError{
				StatusCode: http.StatusConflict,
				Code:       HTTP500InternalServerErrorCode,
			},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := isRetryableServiceError(tc.error)
			if result != tc.expected {
				t.Errorf("isRetryableServiceError(%v) = %v ; wanted %v", tc.error, result, tc.expected)
			}
		})
	}

}
