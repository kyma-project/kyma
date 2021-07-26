module github.com/kyma-project/kyma/tests/application-operator-tests

go 1.16

require (
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210624133846-3e1e71e9f682
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	helm.sh/helm/v3 v3.6.1
	istio.io/client-go v0.0.0-20200814134724-bcbf0ed82b30
	istio.io/gogo-genproto v0.0.0-20200130224810-a0338448499a // indirect
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v0.21.2
	rsc.io/letsencrypt v0.0.3 // indirect
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc92
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/text => golang.org/x/text v0.3.3
)
