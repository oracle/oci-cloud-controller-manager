package framework

import (
	"net/http"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
)

func TestBlockStorageDeleteRetryPolicy(t *testing.T) {
	retryPolicy := blockStorageDeleteRequestMetadata().RetryPolicy
	if retryPolicy == nil {
		t.Fatal("expected a retry policy")
	}
	if retryPolicy.MaximumNumberAttempts == 0 {
		t.Fatal("expected retry attempts to be bounded")
	}

	tests := []struct {
		name       string
		statusCode int
		errorCode  string
		wantRetry  bool
	}{
		{
			name:       "bad gateway",
			statusCode: http.StatusBadGateway,
			errorCode:  "InternalError",
			wantRetry:  true,
		},
		{
			name:       "too many requests",
			statusCode: http.StatusTooManyRequests,
			errorCode:  "TooManyRequests",
			wantRetry:  true,
		},
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			errorCode:  "InvalidParameter",
			wantRetry:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := common.NewOCIOperationResponse(
				nil,
				serviceerror{
					StatusCode: tt.statusCode,
					Code:       tt.errorCode,
					Message:    tt.name,
				},
				1,
			)
			if got := retryPolicy.ShouldRetryOperation(response); got != tt.wantRetry {
				t.Fatalf("expected retry=%t, got %t", tt.wantRetry, got)
			}
		})
	}
}
