package disk

import (
	"reflect"
	"testing"
)

func TestMakeMountArgs(t *testing.T) {
	testCases := []struct {
		name     string
		source   string
		target   string
		fsType   string
		options  []string
		expected []string
	}{
		{
			name:     "mount args for oci-fss with fips",
			source:   "/source",
			target:   "/target",
			fsType:   "oci-fss",
			options:  []string{"fips"},
			expected: []string{"-t", "oci-fss", "-o", "fips", "/source", "/target"},
		},
		{
			name:     "mount args for oci-fss without options",
			source:   "/source",
			target:   "/target",
			fsType:   "oci-fss",
			expected: []string{"-t", "oci-fss", "/source", "/target"},
		},
		{
			name:     "mount args for non encrypted mounts",
			source:   "/source",
			target:   "/target",
			expected: []string{"/source", "/target"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			mountArgs, _ := MakeMountArgs(tt.source, tt.target, tt.fsType, tt.options)
			if !reflect.DeepEqual(mountArgs, tt.expected) {
				t.Errorf("%+v => Expected: %+v Actual: %+v", tt.name, mountArgs, tt.expected)
			}
		})
	}
}
