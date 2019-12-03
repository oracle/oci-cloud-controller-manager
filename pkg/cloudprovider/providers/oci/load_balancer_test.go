package oci

import (
	"reflect"
	"testing"
)

func Test_getDefaultLBSubnets(t *testing.T) {
	type args struct {
		subnet1 string
		subnet2 string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "no default subnets provided",
			args: args{},
			want: []string{""},
		},
		{
			name: "1st subnet provided",
			args: args{"subnet1", ""},
			want: []string{"subnet1"},
		},
		{
			name: "2nd subnet provided",
			args: args{"", "subnet2"},
			want: []string{"", "subnet2"},
		},
		{
			name: "both default subnets provided",
			args: args{"subnet1", "subnet2"},
			want: []string{"subnet1", "subnet2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDefaultLBSubnets(tt.args.subnet1, tt.args.subnet2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDefaultLBSubnets() = %v, want %v", got, tt.want)
			}
		})
	}
}
