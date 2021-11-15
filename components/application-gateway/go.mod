module github.com/kyma-project/kyma/components/application-gateway

go 1.16

require (
	github.com/gorilla/mux v1.8.0
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20211110074047-13002528fca2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v0.22.0
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.11
	github.com/docker/docker => github.com/docker/docker v20.10.8+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc92

	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/text => golang.org/x/text v0.3.3
)
