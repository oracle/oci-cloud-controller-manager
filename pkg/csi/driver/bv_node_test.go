package driver

import "testing"

func Test_getDevicePathAndAttachmentType(t *testing.T) {
	type args struct {
		path []string
	}
	tests := []struct {
		name           string
		args           args
		attachmentType string
		diskByPath     string
		wantErr        bool
	}{
		{
			"Testing PV path with digits only",
			args{path: []string{"/dev/disk/by-path/pci-0000:18:00.0-scsi-0:0:0:5"}},
			"paravirtualized",
			"/dev/disk/by-path/pci-0000:18:00.0-scsi-0:0:0:5",
			false,
		},
		{
			"Testing PV path with hexadecimal controller",
			args{path: []string{"/dev/disk/by-path/pci-0000:1a:00.0-scsi-0:0:4:1"}},
			"paravirtualized",
			"/dev/disk/by-path/pci-0000:1a:00.0-scsi-0:0:4:1",
			false,
		},
		{
			"Testing PV path with hexadecimal Bus",
			args{path: []string{"/dev/disk/by-path/pci-0000:00:ff.0-scsi-0:0:0:1"}},
			"paravirtualized",
			"/dev/disk/by-path/pci-0000:00:ff.0-scsi-0:0:0:1",
			false,
		},
		{
			"Testing ISCSI path",
			args{path: []string{"/dev/disk/by-path/ip-169.254.2.19:3260-iscsi-iqn.2015-12.com.oracleiaas:d0ee92cb-5220-423e-b029-07ae2b2ff08f-lun-1"}},
			"iscsi",
			"/dev/disk/by-path/ip-169.254.2.19:3260-iscsi-iqn.2015-12.com.oracleiaas:d0ee92cb-5220-423e-b029-07ae2b2ff08f-lun-1",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getDevicePathAndAttachmentType(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDevicePathAndAttachmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.attachmentType {
				t.Errorf("getDevicePathAndAttachmentType() got = %v, want attachmentType %v", got, tt.attachmentType)
			}
			if got1 != tt.diskByPath {
				t.Errorf("getDevicePathAndAttachmentType() got1 = %v, want diskByPath %v", got1, tt.diskByPath)
			}
		})
	}
}
