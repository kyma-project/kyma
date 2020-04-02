package finder

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/yaml"

	oldapi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
)

var _ = Describe("Filter", func() {
	Describe("For labels", func() {

		var api1 = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
    name: httpbin-api
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
`)
		var api2 = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
    name: httpbin-api
    labels:
      foo: baz
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
`)
		var api3 = readApi(`
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
    name: httpbin-api
    labels:
      abc: def
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
`)

		It("should correctly filter out by label key and value", func() {
			labelValue := "bar"
			labelsMap := map[string]*string{
				"foo": &labelValue,
			}
			lf := newLabelFilter(labelsMap)
			filtered, reason := lf(api1)
			Expect(filtered).To(BeTrue())
			Expect(reason).To(ContainSubstring("object matches configured label: foo"))

			filtered, reason = lf(api2)
			Expect(filtered).To(BeFalse())
			Expect(reason).To(BeEmpty())

			filtered, reason = lf(api3)
			Expect(filtered).To(BeFalse())
			Expect(reason).To(BeEmpty())

		})

		It("should correctly filter out only by label key", func() {
			labelsMap := map[string]*string{
				"foo": nil,
			}
			lf := newLabelFilter(labelsMap)
			filtered, reason := lf(api1)
			Expect(filtered).To(BeTrue())
			Expect(reason).To(ContainSubstring("object matches configured label: foo"))

			filtered, reason = lf(api2)
			Expect(filtered).To(BeTrue())
			Expect(reason).To(ContainSubstring("object matches configured label: foo"))

			filtered, reason = lf(api3)
			Expect(filtered).To(BeFalse())
			Expect(reason).To(BeEmpty())
		})

		It("should correctly filter out all labelled objects", func() {
			labelValue := "def"
			labelsMap := map[string]*string{
				"foo": nil,
				"abc": &labelValue,
			}
			lf := newLabelFilter(labelsMap)
			filtered, reason := lf(api1)
			Expect(filtered).To(BeTrue())
			Expect(reason).To(ContainSubstring("object matches configured label: foo"))

			filtered, reason = lf(api2)
			Expect(filtered).To(BeTrue())
			Expect(reason).To(ContainSubstring("object matches configured label: foo"))

			filtered, reason = lf(api3)
			Expect(filtered).To(BeTrue())
			Expect(reason).To(ContainSubstring("object matches configured label: abc"))
		})
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
