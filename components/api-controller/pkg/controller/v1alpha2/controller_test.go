package v1alpha2

import "testing"

var fixHostnameTestCases = []struct {
	hostname   string
	domainName string
	expected   string
}{
	{"my-service", "kyma.local", "my-service.kyma.local"},
	{"my-service.kyma.local", "kyma.local", "my-service.kyma.local"},
}

func TestFixHostname(t *testing.T) {
	for _, tc := range fixHostnameTestCases {
		if fixHostname(tc.domainName, tc.hostname) != tc.expected {
			t.Errorf("Fixed hostname %s should be: %s", tc.hostname, tc.expected)
		}
	}
}
