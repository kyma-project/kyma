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
	{"my-service.kyma-local", false},
	{"my-service.my-division", false},
	{"my-service.my-division.kyma.local", false},
	{"my-service.kima.locl", false},
	{"my-service.kyma.local.kyma.local", false},
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
