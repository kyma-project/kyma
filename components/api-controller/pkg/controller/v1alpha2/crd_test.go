package v1alpha2

import (
	"testing"
	"regexp"
)

func TestMatchHostnameWithoutSubdomain(t *testing.T) {
	re := regexp.MustCompile(hostnamePattern)
	hostname := "my-service"
	if !re.MatchString(hostname) {
		t.Error("Hostname should match", hostname)
	}
}

func TestMatchNotHostnameWithSubdomain(t *testing.T) {
	re := regexp.MustCompile(hostnamePattern)
	hostname := "my-service.my-division"
	if re.MatchString(hostname) {
		t.Error("Hostname shouldn't match:", hostname)
	}
}