module github.com/kyma-project/kyma/tests/application-registry-tests

go 1.15

require (
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210624133846-3e1e71e9f682
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
)
