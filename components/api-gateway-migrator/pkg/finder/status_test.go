package finder

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {
	Describe("For status", func() {

		var apiOK = readApi(`
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
status:
  authenticationStatus:
    code: 2
    resource:
      name: httpbin-api
      uid: 3278b6ba-b606-41b2-8708-315f45bdc52a
      version: "166350"
  validationStatus: 2
  virtualServiceStatus:
    code: 2
    resource:
      name: httpbin-api
      uid: e9a7b7c8-fa41-4f4a-b997-e87e1bfc5859
      version: "166380"
`)

		var apiWrongAuthCode = readApi(`
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
status:
  authenticationStatus:
    code: 0
    resource:
      name: httpbin-api
      uid: 3278b6ba-b606-41b2-8708-315f45bdc52a
      version: "166350"
  validationStatus: 2
  virtualServiceStatus:
    code: 2
    resource:
      name: httpbin-api
      uid: e9a7b7c8-fa41-4f4a-b997-e87e1bfc5859
      version: "166380"
`)

		var apiWrongValidationCode = readApi(`
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
status:
  authenticationStatus:
    code: 2
    resource:
      name: httpbin-api
      uid: 3278b6ba-b606-41b2-8708-315f45bdc52a
      version: "166350"
  validationStatus: 0
  virtualServiceStatus:
    code: 2
    resource:
      name: httpbin-api
      uid: e9a7b7c8-fa41-4f4a-b997-e87e1bfc5859
      version: "166380"
`)

		var apiWrongVSCode = readApi(`
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
status:
  authenticationStatus:
    code: 2
    resource:
      name: httpbin-api
      uid: 3278b6ba-b606-41b2-8708-315f45bdc52a
      version: "166350"
  validationStatus: 2
  virtualServiceStatus:
    code: 0
    resource:
      name: httpbin-api
      uid: e9a7b7c8-fa41-4f4a-b997-e87e1bfc5859
      version: "166380"
`)

		It("should correctly filter out objects with invalid status", func() {
			sf := newStatusFilter()

			filtered, reason := sf(apiOK)
			Expect(filtered).To(BeFalse())
			Expect(reason).To(BeEmpty())

			filtered, reason = sf(apiWrongAuthCode)
			Expect(filtered).To(BeTrue())
			Expect(reason).To(ContainSubstring("Invalid authenticationStatus code: 0"))

			filtered, reason = sf(apiWrongValidationCode)
			Expect(filtered).To(BeTrue())
			Expect(reason).To(ContainSubstring("Invalid validationStatus code: 0"))

			filtered, reason = sf(apiWrongVSCode)
			Expect(filtered).To(BeTrue())
			Expect(reason).To(ContainSubstring("Invalid virtualServiceStatus code: 0"))

		})
	})
})
