module github.com/kyma-project/kyma/tests/application-operator-tests

go 1.14

require (
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/godbus/dbus v0.0.0-20190422162347-ade71ed3457e // indirect
	github.com/golangplus/bytes v0.0.0-20160111154220-45c989fe5450 // indirect
	github.com/golangplus/fmt v0.0.0-20150411045040-2a5d6d7d2995 // indirect
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210624133846-3e1e71e9f682
	github.com/opencontainers/runtime-tools v0.0.0-20181011054405-1d69bd0f9c39 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1 // indirect
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	helm.sh/helm/v3 v3.6.1
	istio.io/client-go v0.0.0-20200814134724-bcbf0ed82b30
	istio.io/gogo-genproto v0.0.0-20200130224810-a0338448499a // indirect
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v0.21.2
	k8s.io/kubernetes v1.13.0 // indirect
	vbom.ml/util v0.0.0-20160121211510-db5cfe13f5cc // indirect
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc92
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/text => golang.org/x/text v0.3.3
)
