module github.com/kyma-project/kyma/components/application-operator

go 1.16

require (
	cloud.google.com/go v0.90.0 // indirect
	github.com/docker/docker v20.10.8+incompatible // indirect
	github.com/gorilla/mux v1.8.0
	github.com/kubernetes-sigs/service-catalog v0.3.1
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	helm.sh/helm/v3 v3.6.3
	k8s.io/apimachinery v0.22.0
	k8s.io/cli-runtime v0.22.0
	k8s.io/client-go v0.22.0
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.9.6
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/docker/docker v20.10.8+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc93

	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/text => golang.org/x/text v0.3.3
)
