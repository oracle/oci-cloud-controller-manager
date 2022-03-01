package util

import (
	"errors"
	"testing"
)

func TestGetMetricDimensionForComponent(t *testing.T) {
	var tests = map[string]struct {
		err                     string
		component               string
		expectedMetricDimension string
	}{
		"LB": {
			err:                     "xyz",
			component:               LoadBalancerType,
			expectedMetricDimension: "LB_xyz",
		},
		"CSI": {
			err:                     "xyz",
			component:               CSIStorageType,
			expectedMetricDimension: "CSI_xyz",
		},
		"FVD": {
			err:                     "xyz",
			component:               FVDStorageType,
			expectedMetricDimension: "FVD_xyz",
		},
		"NoError": {
			err:                     "",
			component:               FVDStorageType,
			expectedMetricDimension: "",
		},
		"NoComponent": {
			err:                     "abc",
			component:               "",
			expectedMetricDimension: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actualMetricDimension := GetMetricDimensionForComponent(tt.err, tt.component)
			if actualMetricDimension != tt.expectedMetricDimension {
				t.Errorf("Expected errorType = %s, but got %s", tt.expectedMetricDimension, actualMetricDimension)
				return
			}
		})
	}
}

func TestGetError(t *testing.T) {
	var tests = map[string]struct {
		err           error
		expectedError string
	}{
		"429": {
			err:           errors.New("Service error:InvalidParameter. error foo. http status code: 429. foo"),
			expectedError: Err429,
		},
		"4xx": {
			err:           errors.New("Service error:InvalidParameter. error foo. http status code: 400. foo"),
			expectedError: Err4XX,
		},
		"5xx": {
			err:           errors.New("Service error:InternalError. error bar. http status code: 500. bar"),
			expectedError: Err5XX,
		},
		"LimitError": {
			err:           errors.New("Service error:LimitExceeded. error bar. http status code: 400. foo "),
			expectedError: ErrLimitExceeded,
		},
		"ValidationError": {
			err:           errors.New("foo bar error"),
			expectedError: ErrValidation,
		},
		"NoError": {
			err:           nil,
			expectedError: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actualError := GetError(tt.err)
			if actualError != tt.expectedError {
				t.Errorf("Expected errorType = %s, but got %s", tt.expectedError, actualError)
				return
			}
		})
	}
}
