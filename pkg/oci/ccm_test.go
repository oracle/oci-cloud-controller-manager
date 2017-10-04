package oci

import "testing"

func TestMapProviderIDToInstanceID(t *testing.T) {
	testCases := map[string]struct {
		providerID string
		expected   string
	}{
		"no cloud prefix": {
			providerID: "testid",
			expected:   "testid",
		},
		"cloud prefix": {
			providerID: providerPrefix + "testid",
			expected:   "testid",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := mapProviderIDToInstanceID(tc.providerID)
			if result != tc.expected {
				t.Errorf("Expected instance id %q, but got %q", tc.expected, result)
			}
		})
	}
}
