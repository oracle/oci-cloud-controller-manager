package driver

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
				logger: tt.fields.logger,
			}
			if got := u.getAvailableDomainInNodeLabel(tt.args.fullAD); got != tt.want {
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
			if got := validateFsType(tt.args.logger, tt.args.fsType); got != tt.want {
				t.Errorf("validateFsType() = %v, want %v", got, tt.want)
			}
		})
	}
}
