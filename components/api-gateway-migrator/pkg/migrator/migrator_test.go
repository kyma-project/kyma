package migrator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("randomizeHostnameInFQDN", func() {

	var mockRandomizer = func(length uint) string {
		res := ""
		for i := 0; i < int(length); i++ {
			res += "@"
		}
		return res
	}

	It("should randomize simple hostname", func() {
		hostname := "abcdefghij.kyma.local"
		expected := "@@@@@@ghij.kyma.local"
		actual := randomizeHostnameInFQDN(hostname, mockRandomizer)
		Expect(actual).To(Equal(expected))
	})

	It("should randomize hostname segment if it's 9 characters", func() {
		hostname := "abcdefghi.kyma.local"
		expected := "@@@@@@ghi.kyma.local"
		actual := randomizeHostnameInFQDN(hostname, mockRandomizer)
		Expect(actual).To(Equal(expected))
	})

	It("should randomize hostname segment if it's 8 characters", func() {
		hostname := "abcdefgh.kyma.local"
		expected := "@@@@@@gh.kyma.local"
		actual := randomizeHostnameInFQDN(hostname, mockRandomizer)
		Expect(actual).To(Equal(expected))
	})

	It("should randomize hostname segment if it's 7 characters", func() {
		hostname := "abcdefg.kyma.local"
		expected := "@@@@@@g.kyma.local"
		actual := randomizeHostnameInFQDN(hostname, mockRandomizer)
		Expect(actual).To(Equal(expected))
	})

	It("should replace hostname segment if it's 6 characters", func() {
		hostname := "abcdef.kyma.local"
		expected := "@@@@@@.kyma.local"
		actual := randomizeHostnameInFQDN(hostname, mockRandomizer)
		Expect(actual).To(Equal(expected))
	})

	It("should set random hostname segment of length 6 if it's less than 6 characters", func() {
		hostname := "abcde.kyma.local"
		expected := "@@@@@@.kyma.local"
		actual := randomizeHostnameInFQDN(hostname, mockRandomizer)
		Expect(actual).To(Equal(expected))
	})

	It("should randomize a real hostname", func() {
		hostname := "apimgtst.gke-upgrade-pr-7867-nmnnwmg6sy.a.build.kyma-project.io"
		expected := "@@@@@@st.gke-upgrade-pr-7867-nmnnwmg6sy.a.build.kyma-project.io"
		actual := randomizeHostnameInFQDN(hostname, mockRandomizer)
		Expect(actual).To(Equal(expected))
	})
})
