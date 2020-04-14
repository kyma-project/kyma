package controllers

import (
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
)

func matchKeyValue(key, value string) GomegaMatcher {
	return MatchFields(
		IgnoreExtras,
		Fields{
			"Data": MatchKeys(IgnoreExtras, Keys{
				"test": Equal("me"),
			}),
		},
	)
}

func TestSynchronizeMapEntry(t *testing.T) {
	testCases := []struct {
		desc     string
		cm       corev1.ConfigMap
		expected rebuildImg
		cmMatch  GomegaMatcher
	}{
		{
			desc: "cm contains key with the same value",
			cm: corev1.ConfigMap{
				Data: map[string]string{"test": "me"},
			},
			expected: rebuildImgInessential,
		},
		{
			desc: "cm does not contain key",
			cm: corev1.ConfigMap{
				Data: map[string]string{"other": "key"},
			},
			expected: rebuildImgRequired,
		},
		{
			desc: "cm contains key with different value than expected",
			cm: corev1.ConfigMap{
				Data: map[string]string{"test": "it"},
			},
			expected: rebuildImgRequired,
		},
		{
			desc:     "cm does not contain data",
			cm:       corev1.ConfigMap{},
			expected: rebuildImgRequired,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			key, value := "test", "me"
			g := NewWithT(t)
			actual := synchronizeMapEntry(&tC.cm, key, value)
			g.Expect(actual).To(Equal(tC.expected))
			g.Expect(tC.cm).To(matchKeyValue(key, value))
		})
	}
}
