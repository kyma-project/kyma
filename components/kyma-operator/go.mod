module github.com/kyma-project/kyma/components/kyma-operator

go 1.14

require (
	github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
	github.com/Azure/go-autorest/autorest v0.9.6
	github.com/Azure/go-autorest/autorest/adal v0.8.2
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/hashicorp/go-getter v1.4.1
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/vrischmann/envconfig v1.1.0
	helm.sh/helm/v3 v3.2.1
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/cli-runtime v0.19.7
	k8s.io/client-go v0.19.7
	rsc.io/letsencrypt v0.0.3 // indirect
)
