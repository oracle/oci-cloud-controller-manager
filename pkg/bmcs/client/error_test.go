package client

import (
	"errors"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "search-error-not-found",
			err:      &SearchError{NotFound: true},
			expected: true,
		},
		{
			name:     "search-error-found",
			err:      &SearchError{NotFound: false},
			expected: false,
		},
		{
			name:     "generic-error",
			err:      errors.New("something erroneous"),
			expected: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			isNotFoundErr := IsNotFound(tt.err)
			if isNotFoundErr != tt.expected {
				t.Errorf("IsNotFound(%+v) => %t, wanted %t", tt.err, isNotFoundErr, tt.expected)
			}
		})
	}
}
