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
	"reflect"
	"testing"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"

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
func TestIsCommonTagPresent(t *testing.T) {
	emptyInitialTags := &config.InitialTags{}
	tests := map[string]struct {
		initialtag *config.InitialTags
		want       bool
	}{
		"empty initial tags": {
			initialtag: emptyInitialTags,
			want:       false,
		},
		"No common tags": {
			initialtag: &config.InitialTags{
				LoadBalancer: &config.TagConfig{
					FreeformTags: nil,
					DefinedTags:  nil,
				},
				BlockVolume: &config.TagConfig{
					FreeformTags: nil,
					DefinedTags:  nil,
				},
				FSS: &config.TagConfig{
					FreeformTags: nil,
					DefinedTags:  nil,
				},
			},
			want: false,
		},
		"Common tags": {
			initialtag: &config.InitialTags{
				Common: &config.TagConfig{
					FreeformTags: nil,
					DefinedTags:  nil,
				},
			},
			want: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := IsCommonTagPresent(tc.initialtag)
			if actual != tc.want {
				t.Errorf("Expected %t but got %t", tc.want, actual)
			}
		})
	}
}

func TestMergeTagConfig(t *testing.T) {
	emptyTagConfig := &config.TagConfig{}
	tests := map[string]struct {
		srcTagConfig    *config.TagConfig
		dstTagConfig    *config.TagConfig
		mergedTagConfig config.TagConfig
	}{
		"null test case": {
			srcTagConfig: emptyTagConfig,
			dstTagConfig: emptyTagConfig,
			mergedTagConfig: config.TagConfig{
				FreeformTags: map[string]string{},
				DefinedTags:  map[string]map[string]interface{}{},
			},
		},
		"base test case": {
			srcTagConfig: &config.TagConfig{
				FreeformTags: map[string]string{"foo": "bar"},
				DefinedTags: map[string]map[string]interface{}{
					"ns": {"foo": "bar"},
				},
			},
			dstTagConfig: &config.TagConfig{
				FreeformTags: map[string]string{"foo1": "bar1"},
				DefinedTags: map[string]map[string]interface{}{
					"ns1": {"foo1": "bar1"},
				},
			},
			mergedTagConfig: config.TagConfig{
				FreeformTags: map[string]string{"foo": "bar", "foo1": "bar1"},
				DefinedTags: map[string]map[string]interface{}{
					"ns":  {"foo": "bar"},
					"ns1": {"foo1": "bar1"},
				},
			},
		},
		"test case with conflicting key": {
			srcTagConfig: &config.TagConfig{
				FreeformTags: map[string]string{"foo": "bar"},
				DefinedTags: map[string]map[string]interface{}{
					"ns": {"foo": "bar"},
				},
			},
			dstTagConfig: &config.TagConfig{
				FreeformTags: map[string]string{"foo": "bar1"},
				DefinedTags: map[string]map[string]interface{}{
					"ns": {"foo2": "bar2"},
				},
			},
			mergedTagConfig: config.TagConfig{
				FreeformTags: map[string]string{"foo1": "bar1"},
				DefinedTags: map[string]map[string]interface{}{
					"ns": {"foo2": "bar2"},
				},
			},
		},
		"test case with one empty config - 1": {
			srcTagConfig: emptyTagConfig,
			dstTagConfig: &config.TagConfig{
				FreeformTags: map[string]string{"foo": "bar1"},
				DefinedTags: map[string]map[string]interface{}{
					"ns": {"foo2": "bar2"},
				},
			},
			mergedTagConfig: config.TagConfig{
				FreeformTags: map[string]string{"foo": "bar1"},
				DefinedTags: map[string]map[string]interface{}{
					"ns": {"foo2": "bar2"},
				},
			},
		},
		"test case with one empty config - 2": {
			srcTagConfig: &config.TagConfig{
				FreeformTags: map[string]string{"foo": "bar1"},
				DefinedTags: map[string]map[string]interface{}{
					"ns": {"foo2": "bar2"},
				},
			},
			dstTagConfig: emptyTagConfig,
			mergedTagConfig: config.TagConfig{
				FreeformTags: map[string]string{"foo": "bar1"},
				DefinedTags: map[string]map[string]interface{}{
					"ns": {"foo2": "bar2"},
				},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := MergeTagConfig(tc.srcTagConfig, tc.dstTagConfig)
			if !reflect.DeepEqual(actual.FreeformTags, tc.mergedTagConfig.FreeformTags) &&
				!reflect.DeepEqual(actual.DefinedTags, tc.mergedTagConfig.DefinedTags) {
				t.Errorf("Expected %v but got %v", tc.mergedTagConfig, actual)
			}
		})
	}
}
