module github.com/kyma-project/kyma/components/kyma-operator

go 1.14

require (
	github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.18
	github.com/Azure/go-autorest/autorest/adal v0.9.13
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/hashicorp/go-getter v1.4.1
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/vrischmann/envconfig v1.1.0
	helm.sh/helm/v3 v3.5.2
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.20.2
	k8s.io/client-go v0.20.2
)

replace (
	// There is a problem with Helm in version 3.5.2 - see more: https://github.com/helm/helm/issues/9354
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	golang.org/x/crypto/ssh => golang.org/x/crypto/ssh v0.0.0-20201216223049-8b5274cf687f
)
