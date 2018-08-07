package v1alpha2

import (
	"testing"
	"regexp"
)

func TestMatchSimpleHostname(t *testing.T) {
	re := regexp.MustCompile(hostnamePattern("kyma.local"))
	hostname := "my-service"
	if !re.MatchString(hostname) {
		t.Error("Hostname should match:", hostname)
	}
}

func TestMatchCompleteHostname(t *testing.T) {
	re := regexp.MustCompile(hostnamePattern("kyma.local"))
	hostname := "my-service.kyma.local"
	if !re.MatchString(hostname) {
		t.Error("Hostname should match:", hostname)
	}
}

func TestMatchNotSimpleHostnameWithSubdomain(t *testing.T) {
	re := regexp.MustCompile(hostnamePattern("kyma.local"))
	hostname := "my-service.my-division"
	if re.MatchString(hostname) {
		t.Error("Hostname shouldn't match:", hostname)
	}
}

func TestMatchNotCompleteHostnameWithSubdomain(t *testing.T) {
	re := regexp.MustCompile(hostnamePattern("kyma.local"))
	hostname := "my-service.my-division.kyma.local"
	if re.MatchString(hostname) {
		t.Error("Hostname shouldn't match:", hostname)
	}
}

func TestMatchNotCompleteHostnameWithDifferentDomain(t *testing.T) {
	re := regexp.MustCompile(hostnamePattern("kyma.local"))
	hostname := "my-service.my-division.kima.locl"
	if re.MatchString(hostname) {
		t.Error("Hostname shouldn't match:", hostname)
	}
}