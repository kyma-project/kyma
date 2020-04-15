module github.com/kyma-project/kyma/components/api-gateway-migrator

go 1.13

require (
	github.com/avast/retry-go v2.4.1+incompatible
	github.com/kr/pretty v0.2.0 // indirect
	github.com/kyma-incubator/api-gateway v0.0.0-20200228120303-0c4d7ae5fdaf
	github.com/kyma-project/kyma v0.5.1-0.20200109154037-119010d0810e
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/ory/oathkeeper-maester v0.0.2-beta.1
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v1.6.4 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	knative.dev/pkg v0.0.0-20190807140856-4707aad818fe
	sigs.k8s.io/controller-runtime v0.2.0-beta.4
)
