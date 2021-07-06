// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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

package logging

import (
	"reflect"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestFieldsFromEnv(t *testing.T) {
	testCases := map[string]struct {
		env    []string
		fields []zapcore.Field
	}{
		"single": {
			env: []string{"LOG_FIELD_foo=bar"},
			fields: []zapcore.Field{
				zap.String("foo", "bar"),
			},
		},
		"multiple": {
			env: []string{
				"LOG_FIELD_foo=bar",
				"LOG_FIELD_bar=baz",
			},
			fields: []zapcore.Field{
				zap.String("foo", "bar"),
				zap.String("bar", "baz"),
			},
		},
		"handles_equals_in_value": {
			env: []string{
				"LOG_FIELD_foo=a=b",
			},
			fields: []zapcore.Field{
				zap.String("foo", "a=b"),
			},
		},
		"handles_empty_value": {
			env: []string{
				"LOG_FIELD_foo=",
			},
			fields: []zapcore.Field{
				zap.String("foo", ""),
			},
		},
		"ignores_non_log_field": {
			env: []string{
				"LOG_FIELD_foo=bar",
				"NOT_A_FIELD=1",
			},
			fields: []zapcore.Field{
				zap.String("foo", "bar"),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			fields := FieldsFromEnv(tc.env)
			if !reflect.DeepEqual(fields, tc.fields) {
				t.Errorf("Got incorrect fields:\nexpected=%+v\nactual=%+v", tc.fields, fields)
			}
		})
	}
}
