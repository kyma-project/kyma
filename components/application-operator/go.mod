module github.com/kyma-project/kyma/components/application-operator

go 1.14

require (
	cloud.google.com/go v0.65.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/gorilla/mux v1.8.0
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.1.0
	helm.sh/helm/v3 v3.5.3
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.6.2
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/deislabs/oras => github.com/deislabs/oras v0.10.0
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc7

	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/text => golang.org/x/text v0.3.3

	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.2
)
