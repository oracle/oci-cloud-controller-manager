package main

import (
	"os"
	"testing"
)

func Test_IsLustreDriverEnabled(t *testing.T) {
	tests := []struct {
		envValue string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"TrUe", true},
		{"false", false},
		{"", false},
		{"random", false},
	}

	for _, tc := range tests {
		// Set or unset the environment variable based on the test case.
		if tc.envValue == "" {
			os.Unsetenv("LUSTRE_DRIVER_ENABLED")
		} else {
			os.Setenv("LUSTRE_DRIVER_ENABLED", tc.envValue)
		}

		// Our logic under test: compare the environment variable with "true" (case-insensitive).
		enableLustreDriver := IsLustreDriverEnabled()

		if enableLustreDriver != tc.expected {
			t.Errorf("For LUSTRE_DRIVER_ENABLED=%q, expected %v but got %v",
				tc.envValue, tc.expected, enableLustreDriver)
		}
	}
}
