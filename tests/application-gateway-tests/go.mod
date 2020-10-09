module github.com/kyma-project/kyma/tests/application-gateway-tests

go 1.14

require (
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20200903161647-0fae3728c173
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.2.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.1
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc92
	golang.org/x/text => golang.org/x/text v0.3.3
)
