package csi_util

import "testing"

func TestValidateLustreVolumeId(t *testing.T) {
	tests := []struct {
		input                    string
		expectedValidationResult bool
		expectedLnetLable        string
	}{
		// Valid cases
		{"192.168.227.11@tcp1:192.168.227.12@tcp1:/demo", true, "tcp1"},
		{"192.168.227.11@tcp1:/demo", true,"tcp1"},
		// Invalid cases
		{"192.168.227.11@tcp1:192.168.227.12@tcp1", false,"tcp1"}, // No fsname provided
		{"192.168.227.11@tcp1:192.168.227.12@tcp1:demo", false,"tcp1"}, // fsname not starting with "/"
		{"invalidip@tcp1:192.168.227.12@tcp1:/demo", false,""},  // Invalid IP address
		{"192.168.227.11@tcp1:invalidip@tcp1:/demo", false,"tcp1"},  // Invalid IP address
		{"192.168.227.11@:192.168.227.12@:tcp1/demo", false, ""}, // No Lnet label provided
		{"192.168.227.11@tcp1:192.168.227.12:/demo", false, "tcp1"}, // No Lnet label provided in 2nd MGS NID
		// Empty input
		{"", false,""},

		// Single IP
		{"192.168.227.11", false,""}, // Missing ":" in the input
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			validationResult, lnetLabel := ValidateLustreVolumeId(test.input)
			if validationResult != test.expectedValidationResult || lnetLabel != test.expectedLnetLable {
				t.Errorf("For input '%s', expected validationResult : %v & lnetLable : %v but got validationResult : %v & lnetLable : %v", test.input, test.expectedValidationResult, test.expectedLnetLable, validationResult, lnetLabel)
			}
		})
	}
}
