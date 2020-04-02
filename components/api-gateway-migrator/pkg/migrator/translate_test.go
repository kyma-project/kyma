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
	It("should translate simple api", func() {

		var api = `
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
`

		var expectedApiRuleYaml = `
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
    - path: /favicon.ico
      methods: ["GET", "PUT", "POST", "DELETE"]
      accessStrategies:
        - handler: allow
    - path: /anything.+
      methods: ["GET", "PUT", "POST", "DELETE"]
      accessStrategies:
        - handler: allow
    - path: /.*
      methods: ["GET", "PUT", "POST", "DELETE"]
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers:
              - "https://dex.kyma.local"
            jwks_urls:
              - "http://dex-service.kyma-system.svc.cluster.local:5556/keys"
`

		apiReader := strings.NewReader(api)
		buffSize := 1000
		apiObj := oldapi.Api{}
		err := yaml.NewYAMLOrJSONDecoder(apiReader, buffSize).Decode(&apiObj)
		Expect(err).To(BeNil())

		apiRuleReader := strings.NewReader(expectedApiRuleYaml)
		apiRuleObj := newapi.APIRule{}
		yaml.NewYAMLOrJSONDecoder(apiRuleReader, buffSize).Decode(&apiRuleObj)
		//isExternal := false
		//apiRuleObj.Spec.Service.IsExternal = &isExternal

		actual := translateToApiRule(&apiObj)
		Expect(actual.Status).To(BeEquivalentTo(apiRuleObj.Status))
		Expect(actual.Spec).To(BeEquivalentTo(apiRuleObj.Spec))
		Expect(actual.TypeMeta).To(BeEquivalentTo(apiRuleObj.TypeMeta))
		Expect(actual.ObjectMeta).To(BeEquivalentTo(apiRuleObj.ObjectMeta))
	})
})
