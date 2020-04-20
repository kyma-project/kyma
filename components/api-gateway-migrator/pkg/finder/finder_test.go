package finder

import (
	oldapi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Finder", func() {
	Describe("Object", func() {

		var api1 = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
  name: api1
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

		var api2 = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
  name: api2
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

		var api3 = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
    name: api3
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

		var api4 = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
    name: api4
    labels:
      foo: bar
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

		var api5 = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
  name: api5
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

		It("should remove all invalid objects", func() {
			labelValue := "bar"
			labelsMap := map[string]*string{
				"foo": &labelValue,
			}
			fnd := New(nil, labelsMap)

			apiList := []oldapi.Api{*api1, *api2, *api3, *api4, *api5}
			filtered := fnd.filterApis(apiList)
			Expect(len(filtered)).To(Equal(2))
			Expect(filtered[0].Name).To(Equal("api1"))
			Expect(filtered[1].Name).To(Equal("api5"))
		})
	})
})
