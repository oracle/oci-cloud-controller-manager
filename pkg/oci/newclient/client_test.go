package client

import (
	"testing"

	"github.com/oracle/oci-go-sdk/core"
)

func TestInstanceTerminalState(t *testing.T) {
	testCases := map[string]struct {
		state    core.InstanceLifecycleStateEnum
		expected bool
	}{
		"not terminal - running": {
			state:    core.InstanceLifecycleStateRunning,
			expected: false,
		},
		"not terminal - stopped": {
			state:    core.InstanceLifecycleStateStopped,
			expected: false,
		},
		"is terminal - terminating": {
			state:    core.InstanceLifecycleStateTerminating,
			expected: true,
		},
		"is terminal - terminated": {
			state:    core.InstanceLifecycleStateTerminated,
			expected: true,
		},
		"is terminal - unknown": {
			state:    "UNKNOWN",
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := IsInstanceInTerminalState(&core.Instance{
				LifecycleState: tc.state,
			})
			if result != tc.expected {
				t.Errorf("IsInstanceInTerminalState(%q) = %v ; wanted %v", tc.state, result, tc.expected)
			}
		})
	}
}
