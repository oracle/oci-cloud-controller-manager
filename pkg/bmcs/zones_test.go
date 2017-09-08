package bmcs

import (
	"testing"
)

var mapAvailabilityDomainToFailureDomainTestCases = []struct {
	ad string
	fd string
}{
	{ad: "NWuj:PHX-AD-1", fd: "PHX-AD-1"},
	{ad: "NWuj:PHX-AD-2", fd: "PHX-AD-2"},
	{ad: "NWuj:PHX-AD-3", fd: "PHX-AD-3"},
	{ad: "", fd: ""},
	{ad: "PHX-AD-3", fd: "PHX-AD-3"},
}

func TestMapAvailabilityDomainToFailureDomain(t *testing.T) {
	for _, tt := range mapAvailabilityDomainToFailureDomainTestCases {
		v := mapAvailabilityDomainToFailureDomain(tt.ad)
		if v != tt.fd {
			t.Errorf("mapAvailabilityDomainToFailureDomain(%q) => %q, want %q", tt.ad, v, tt.fd)
		}
	}
}
