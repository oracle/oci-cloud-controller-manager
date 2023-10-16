// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"errors"
	"testing"

	errors2 "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
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
		"429_UpperCase": {
			err:           errors.New("Service error:InvalidParameter. error foo. Http Status Code: 429. foo"),
			expectedError: Err429,
		},
		"4xx": {
			err:           errors.New("Service error:InvalidParameter. error foo. http status code: 400. foo"),
			expectedError: Err4XX,
		},
		"4xx_UpperCase": {
			err:           errors.New("Service error:InvalidParameter. error foo. Http Status Code: 400. foo"),
			expectedError: Err4XX,
		},
		"5xx": {
			err:           errors.New("Service error:InternalError. error bar. http status code: 500. bar"),
			expectedError: Err5XX,
		},
		"5xx_UpperCase": {
			err:           errors.New("Service error:InternalError. error bar. Http Status Code: 500. bar"),
			expectedError: Err5XX,
		},
		"LimitError_AsServiceError": {
			err:           errors.New("Service error:LimitExceeded. error bar. http status code: 400. foo "),
			expectedError: ErrLimitExceeded,
		},
		"LimitError_AsErrorCode": {
			err:           errors.New("Http Status Code: 400. Error Code: LimitExceeded. Opc request id: "),
			expectedError: ErrLimitExceeded,
		},
		"ContextTimeoutError": {
			err:           errors2.Wrap(errors2.Wrap(errors2.WithStack(wait.ErrWaitTimeout), "Bar"), "Foo"),
			expectedError: ErrCtxTimeout,
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
