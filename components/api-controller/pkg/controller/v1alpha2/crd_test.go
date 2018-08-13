package v1alpha2

import (
	"regexp"
	"testing"
)

var hostnameCases = []struct {
	hostname string
	valid    bool
}{
	{"my-service", true},
	{"my-service.kyma.local", true},
	{"dot-is-escaped.kyma-local", false},
	{"with-subdomain.my-division", false},
	{"with-subdomain.my-division.kyma.local", false},
	{"wrong-domain.kima.locl", false},
	{"duplicated-domain.kyma.local.kyma.local", false},
	{"my-special-very-too-long-hostname-that-is-not-compliant-with-relevant-rfc", false},
	{"my-special-very-too-long-hostname-that-is-not-compliant-with-relevant-rfc.kyma.local", false},
}

func TestMatchHostname(t *testing.T) {
	re := regexp.MustCompile(hostnamePattern("kyma.local"))
	for _, tc := range hostnameCases {
		t.Run(tc.hostname, func(t *testing.T) {
			if re.MatchString(tc.hostname) != tc.valid {
				t.Errorf("Hostname '%s' should match: %v", tc.hostname, tc.valid)
			}
		})
	}
}
