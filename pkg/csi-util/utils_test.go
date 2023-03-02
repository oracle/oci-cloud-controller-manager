// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
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

package csi_util

import (
	"testing"

	"go.uber.org/zap"
)

func TestUtil_getAvailableDomainInNodeLabel(t *testing.T) {
	type fields struct {
		logger *zap.SugaredLogger
	}
	type args struct {
		fullAD string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Get AD name from the tenancy specific AD name.",
			fields: fields{
				logger: zap.S(),
			},
			args: args{fullAD: "zkJl:US-ASHBURN-AD-1"},
			want: "US-ASHBURN-AD-1",
		},
		{
			name: "Get AD name from the tenancy specific AD name for empty string",
			fields: fields{
				logger: zap.S(),
			},
			args: args{fullAD: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &Util{
				Logger: tt.fields.logger,
			}
			if got := u.GetAvailableDomainInNodeLabel(tt.args.fullAD); got != tt.want {
				t.Errorf("Util.getAvailableDomainInNodeLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateFsType(t *testing.T) {
	type args struct {
		logger *zap.SugaredLogger
		fsType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Return ext4",
			args: args{
				logger: zap.S(),
				fsType: "ext4",
			},
			want: "ext4",
		},
		{
			name: "Return ext3",
			args: args{
				logger: zap.S(),
				fsType: "ext3",
			},
			want: "ext3",
		},
		{
			name: "Return xfs",
			args: args{
				logger: zap.S(),
				fsType: "xfs",
			},
			want: "xfs",
		},
		{
			name: "Return default ext4 for empty string",
			args: args{
				logger: zap.S(),
				fsType: "",
			},
			want: "ext4",
		},
		{
			name: "Return default ext4 for unsupported string",
			args: args{
				logger: zap.S(),
				fsType: "xxxxx",
			},
			want: "ext4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateFsType(tt.args.logger, tt.args.fsType); got != tt.want {
				t.Errorf("validateFsType() = %v, want %v", got, tt.want)
			}
		})
	}
}
