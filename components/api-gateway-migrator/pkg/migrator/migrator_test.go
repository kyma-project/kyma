package migrator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("shortenHostName", func() {

	It("should shorten simple hostname", func() {
		hostname := "abcdefghij.kyma.local"
		expected := "aij.kyma.local"
		actual := shortenHostName(hostname, 7)
		Expect(actual).To(Equal(expected))
	})

	It("should shorten hostname segment if it's two chars longer than required", func() {
		hostname := "abcdefghi.kyma.local"
		expected := "ai.kyma.local"
		actual := shortenHostName(hostname, 7)
		Expect(actual).To(Equal(expected))
	})

	It("should remove hostname segment if it's only one char longer than required", func() {
		hostname := "abcdefgh.kyma.local"
		expected := "kyma.local"
		actual := shortenHostName(hostname, 7)
		Expect(actual).To(Equal(expected))
	})

	It("should remove hostname segment if it's length is equal to required", func() {
		hostname := "abcdefg.kyma.local"
		expected := "kyma.local"
		actual := shortenHostName(hostname, 7)
		Expect(actual).To(Equal(expected))
	})

	It("should remove a short hostname segment", func() {
		hostname := "ab.cdefgh.kyma.local"
		expected := "ch.kyma.local"
		actual := shortenHostName(hostname, 7)
		Expect(actual).To(Equal(expected))
	})

	It("should remove two short hostname segments", func() {
		hostname := "ab.cd.efgh.kyma.local"
		expected := "egh.kyma.local"
		actual := shortenHostName(hostname, 7)
		Expect(actual).To(Equal(expected))
	})

	It("should remove two short hostname segments", func() {
		hostname := "abc.def.ghij.kyma.local"
		expected := "ghij.kyma.local"
		actual := shortenHostName(hostname, 7)
		Expect(actual).To(Equal(expected))
	})
})
