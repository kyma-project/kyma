package util

import "testing"

func TestGetKnSubscriptionName(t *testing.T) {
	testCases := []struct {
		name     string
		ns       string
		kymaSub  string
		expected string
	}{{
		name:     "A kyma sub with period",
		ns:       "test",
		kymaSub:  "foo.bar",
		expected: "foo.bar--test",
	}, {
		name:     "A kyma sub with hyphen with",
		ns:       "test",
		kymaSub:  "foo-bar",
		expected: "foo-bar--test",
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetKnSubscriptionName(&tc.kymaSub, &tc.ns)
			if got != tc.expected {
				t.Errorf("Invalid Knative Subscription, Expected: %s, Got: %s", tc.expected, got)
			}
		})
	}
	kymaSub := "ky-sub"
	kymaNs := "kyma-ns"
	expectedKnSub := "ky-sub--kyma-dns"
	knSubName := GetKnSubscriptionName(&kymaSub, &kymaNs)
	if knSubName != expectedKnSub {
		t.Fatalf("Invalid Knative subscription name, Expected: %s, Got: %s", expectedKnSub, knSubName)
	}
}

func TestEscapeHyphensAndPeriods(t *testing.T) {
	testCases := []struct {
		name     string
		sample   string
		expected string
	}{{
		name:     "A string with period should have a hyphen with p",
		sample:   "foo.bar",
		expected: "foo-pbar",
	}, {
		name:     "A string with hyphen should have a hyphen with d",
		sample:   "foo-bar",
		expected: "foo-dbar",
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := escapeHyphensAndPeriods(&tc.sample)
			if got != tc.expected {
				t.Errorf("Invalid conversion of hyphens and periods, Expected: %s, Got: %s", tc.expected, got)
			}
		})
	}

}
