module github.com/kyma-project/kyma/components/application-connectivity-validator

go 1.16

require (
	github.com/gorilla/mux v1.8.0
	github.com/kyma-project/kyma/common/logging v0.0.0-20210318081026-665ca4cda3f6
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20211110074047-13002528fca2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/zap v1.18.1
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v0.22.0
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.11
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc92

	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/text => golang.org/x/text v0.3.3

	k8s.io/api => k8s.io/api v0.0.0-20201020200614-54bcc311e327
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20201020200440-554eef9dbf66
	k8s.io/client-go => k8s.io/client-go v0.0.0-20201020200834-d1a4fe5f2d96
)
