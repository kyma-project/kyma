package finder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {
	Describe("For object structure", func() {

		var apiOk1 = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
  name: httpbin-api
spec:
  service:
    name: httpbin
    port: 8000
  hostname: httpbin.kyma.local
  authentication:
  - type: JWT
    jwt:
      issuer: https://dex.kyma.local
      jwksUri: http://dex-service.kyma-system.svc.cluster.local:5556/keys
  - type: JWT
    jwt:
      issuer: https://kex.kyma.local
      jwksUri: http://kex-service.kyma-system.svc.cluster.local:5556/keys
`)

		var apiOk2 = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
  name: httpbin-api
spec:
  service:
    name: httpbin
    port: 8000
  hostname: httpbin.kyma.local
  authentication:
  - type: JWT
    jwt:
      issuer: https://dex.kyma.local
      jwksUri: http://dex-service.kyma-system.svc.cluster.local:5556/keys
      triggerRule:
        excludedPaths:
        - suffix: /favicon.ico
        - regex: /anything.+
  - type: JWT
    jwt:
      issuer: https://kex.kyma.local
      jwksUri: http://kex-service.kyma-system.svc.cluster.local:5556/keys
      triggerRule:
        excludedPaths:
        - regex: /anything.+
        - suffix: /favicon.ico
`)

		var apiWrong = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
  name: httpbin-api
spec:
  service:
    name: httpbin
    port: 8000
  hostname: httpbin.kyma.local
  authentication:
  - type: JWT
    jwt:
      issuer: https://dex.kyma.local
      jwksUri: http://dex-service.kyma-system.svc.cluster.local:5556/keys
      triggerRule:
        excludedPaths:
        - regex: /anything.+
        - prefix: /favicon.ico
  - type: JWT
    jwt:
      issuer: https://kex.kyma.local
      jwksUri: http://kex-service.kyma-system.svc.cluster.local:5556/keys
      triggerRule:
        excludedPaths:
        - regex: /anything.+
        - suffix: /favicon.ico
`)

		var filter = newJwtStructureFilter()

		It("should allow for objects with two JWT elements without excludePaths", func() {
			filtered, reason := filter(apiOk1)
			Expect(filtered).To(BeFalse())
			Expect(reason).To(BeEmpty())
		})

		It("should allow for objects with two JWT elements with same excludePaths", func() {
			filtered, reason := filter(apiOk2)
			Expect(filtered).To(BeFalse())
			Expect(reason).To(BeEmpty())
		})

		It("should filter out objects with two JWT elements and different excludePaths", func() {
			filtered, reason := filter(apiWrong)
			Expect(filtered).To(BeTrue())
			Expect(reason).To(ContainSubstring("object is configured with more than one jwt authentication that contain different triggerRules"))
		})
	})
})
