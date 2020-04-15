package migrator

import (
	"strings"

	newapi "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/yaml"

	oldapi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
)

var _ = Describe("translate func", func() {
	var gateway = "kyma-gateway.kyma-system.svc.cluster.local"

	It("should translate simple api", func() {

		var inputApi = readApi(`
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
`)

		var expected = readApiRule(`
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: httpbin-api
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  service:
    name: httpbin
    port: 8000
    host: httpbin.kyma.local
    external: false
  rules:
    - path: .*/favicon.ico
      methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
      accessStrategies:
        - handler: allow
    - path: /anything.+
      methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
      accessStrategies:
        - handler: allow
    - path: /.*
      methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers:
              - "https://dex.kyma.local"
            jwks_urls:
              - "http://dex-service.kyma-system.svc.cluster.local:5556/keys"
`)

		actual, err := translateToApiRule(inputApi, gateway)

		Expect(err).To(BeNil())
		Expect(actual.Status).To(BeEquivalentTo(expected.Status))
		Expect(actual.Spec).To(BeEquivalentTo(expected.Spec))
		Expect(actual.TypeMeta).To(BeEquivalentTo(expected.TypeMeta))
		Expect(actual.ObjectMeta).To(BeEquivalentTo(expected.ObjectMeta))
	})

	It("should translate api with many excludePaths", func() {

		var inputApi = readApi(`
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
            - prefix: /pref/
            - suffix: /suffix.ico
            - regex: /anything.*
            - exact: /exact/path/to/resource.jpg
`)

		var expected = readApiRule(`
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: httpbin-api
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  service:
    name: httpbin
    port: 8000
    host: httpbin.kyma.local
    external: false
  rules:
  - accessStrategies:
    - handler: allow
    methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
    path: /pref/.*
  - accessStrategies:
    - handler: allow
    methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
    path: .*/suffix.ico
  - accessStrategies:
    - handler: allow
    methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
    path: /anything.*
  - accessStrategies:
    - handler: allow
    methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
    path: /exact/path/to/resource.jpg
  - accessStrategies:
    - config:
        jwks_urls:
        - http://dex-service.kyma-system.svc.cluster.local:5556/keys
        trusted_issuers:
        - https://dex.kyma.local
      handler: jwt
    methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
    path: /.*
`)
		actual, err := translateToApiRule(inputApi, gateway)

		Expect(err).To(BeNil())
		Expect(actual.Status).To(BeEquivalentTo(expected.Status))
		Expect(actual.Spec).To(BeEquivalentTo(expected.Spec))
		Expect(actual.TypeMeta).To(BeEquivalentTo(expected.TypeMeta))
		Expect(actual.ObjectMeta).To(BeEquivalentTo(expected.ObjectMeta))
	})

	It("should translate simple api with 3 authentications", func() {

		var inputApi = readApi(`
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
        issuer: https://keks.kyma.local
        jwksUri: http://keks-service.kyma-system.svc.cluster.local:5556/keys
    - type: JWT
      jwt:
        issuer: https://peks.kyma.local
        jwksUri: http://peks-service.kyma-system.svc.cluster.local:5556/keys
`)

		var expected = readApiRule(`
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: httpbin-api
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  service:
    name: httpbin
    port: 8000
    host: httpbin.kyma.local
    external: false
  rules:
  - accessStrategies:
    - config:
        jwks_urls:
        - http://dex-service.kyma-system.svc.cluster.local:5556/keys
        - http://keks-service.kyma-system.svc.cluster.local:5556/keys
        - http://peks-service.kyma-system.svc.cluster.local:5556/keys
        trusted_issuers:
        - https://dex.kyma.local
        - https://keks.kyma.local
        - https://peks.kyma.local
      handler: jwt
    methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
    path: /.*
`)
		actual, err := translateToApiRule(inputApi, gateway)

		//Expect(actual).To(BeEquivalentTo(expected))
		Expect(err).To(BeNil())
		Expect(actual.Status).To(BeEquivalentTo(expected.Status))
		Expect(actual.Spec).To(BeEquivalentTo(expected.Spec))
		Expect(actual.TypeMeta).To(BeEquivalentTo(expected.TypeMeta))
		Expect(actual.ObjectMeta).To(BeEquivalentTo(expected.ObjectMeta))
	})

	It("should translate simple api with 2 authentications and identical excludePaths", func() {

		var inputApi = readApi(`
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
          - prefix: /images/
    - type: JWT
      jwt:
        issuer: https://dex.nightly.a.build.kyma-project.io
        jwksUri: https://dex.nightly.a.build.kyma-project.io/keys
        triggerRule:
          excludedPaths:
          - prefix: /images/
`)

		var expected = readApiRule(`
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: httpbin-api
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  service:
    name: httpbin
    port: 8000
    host: httpbin.kyma.local
    external: false
  rules:
  - path: /images/.*
    methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
    accessStrategies:
    - handler: allow
  - path: /.*
    methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
    accessStrategies:
    - handler: jwt
      config:
        jwks_urls:
        - http://dex-service.kyma-system.svc.cluster.local:5556/keys
        - https://dex.nightly.a.build.kyma-project.io/keys
        trusted_issuers:
        - https://dex.kyma.local
        - https://dex.nightly.a.build.kyma-project.io
`)

		actual, err := translateToApiRule(inputApi, gateway)

		Expect(err).To(BeNil())
		Expect(actual.Status).To(BeEquivalentTo(expected.Status))
		Expect(actual.Spec).To(BeEquivalentTo(expected.Spec))
		Expect(actual.TypeMeta).To(BeEquivalentTo(expected.TypeMeta))
		Expect(actual.ObjectMeta).To(BeEquivalentTo(expected.ObjectMeta))
	})

	It("should translate api with no authentications defined", func() {

		var inputApi = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
  name: httpbin-api
spec:
  service:
    name: httpbin
    port: 8000
  hostname: httpbin.kyma.local
  authentication: []
`)

		var expected = readApiRule(`
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: httpbin-api
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  service:
    name: httpbin
    port: 8000
    host: httpbin.kyma.local
    external: false
  rules:
  - path: /.*
    methods: ["GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"]
    accessStrategies:
    - handler: allow
`)

		actual, err := translateToApiRule(inputApi, gateway)

		Expect(err).To(BeNil())
		Expect(actual.Status).To(BeEquivalentTo(expected.Status))
		Expect(actual.Spec).To(BeEquivalentTo(expected.Spec))
		Expect(actual.TypeMeta).To(BeEquivalentTo(expected.TypeMeta))
		Expect(actual.ObjectMeta).To(BeEquivalentTo(expected.ObjectMeta))
	})
})

func readApi(yamlValue string) *oldapi.Api {
	defer GinkgoRecover()

	apiReader := strings.NewReader(yamlValue)
	buffSize := 1000
	apiObj := oldapi.Api{}
	err := yaml.NewYAMLOrJSONDecoder(apiReader, buffSize).Decode(&apiObj)
	if err != nil {
		Fail(err.Error())
	}
	return &apiObj
}

func readApiRule(yamlValue string) *newapi.APIRule {
	defer GinkgoRecover()

	apiReader := strings.NewReader(yamlValue)
	buffSize := 1000
	apiObj := newapi.APIRule{}
	err := yaml.NewYAMLOrJSONDecoder(apiReader, buffSize).Decode(&apiObj)
	if err != nil {
		Fail(err.Error())
	}
	return &apiObj
}
