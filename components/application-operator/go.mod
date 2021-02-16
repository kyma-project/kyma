module github.com/kyma-project/kyma/components/application-operator

go 1.14

require (
	cloud.google.com/go v0.65.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/gorilla/mux v1.8.0
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.1.0
	gotest.tools/v3 v3.0.2 // indirect
	helm.sh/helm/v3 v3.3.4
	k8s.io/apimachinery v0.18.15
	k8s.io/cli-runtime v0.18.15
	k8s.io/client-go v0.18.15
	k8s.io/klog v1.0.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.6.2
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.3
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20210128214336-420b1d36250f+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc93

	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/text => golang.org/x/text v0.3.3
)
