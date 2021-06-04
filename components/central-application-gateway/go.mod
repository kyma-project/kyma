module github.com/kyma-project/kyma/components/central-application-gateway

go 1.14

require (
	github.com/gorilla/mux v1.8.0
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210604090801-9a75cdfd1c7c
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.1
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc92
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/text => golang.org/x/text v0.3.3
)
