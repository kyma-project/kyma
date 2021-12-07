module github.com/kyma-project/kyma/components/central-application-connectivity-validator

go 1.16

require (
	github.com/gorilla/mux v1.8.0
	github.com/kyma-project/kyma/common/logging v0.0.0-20210318081026-665ca4cda3f6
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210624133846-3e1e71e9f682
	github.com/oklog/run v1.1.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.1.0
	go.uber.org/zap v1.18.1
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/controller-runtime v0.9.6
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.11
	github.com/docker/docker => github.com/docker/docker v20.10.8+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc92
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20211115234514-b4de73f9ece8
	golang.org/x/text => golang.org/x/text v0.3.3
)
