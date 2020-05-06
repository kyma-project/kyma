package kubernetes

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
)

var _ = ginkgo.Describe("Config", func() {
	ginkgo.It("Excluded namespaces should not have length of 1", func() {
		// this test is just to be secure if someone used "," instead of ";"
		// I assume that we have more than 1 namespace we need to exclude by default
		
		cfg := Config{}
		err := envconfig.Init(&cfg)
		gomega.Expect(err).To(gomega.Succeed())

		gomega.Expect(cfg.ExcludedNamespaces).NotTo(gomega.HaveLen(1))
	})
})
