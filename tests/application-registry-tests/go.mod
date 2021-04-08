module github.com/kyma-project/kyma/tests/application-registry-tests

go 1.15

require (
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20201126144127-ef3483539864
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/kyma-project/kyma/components/application-operator => github.com/kyma-project/kyma/components/application-operator v0.0.0-20200610105106-1066324c83cd
	k8s.io/api => k8s.io/api v0.17.14
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.14
	k8s.io/client-go => k8s.io/client-go v0.17.14
)
