module github.com/kyma-project/kyma/tests/application-operator-tests

go 1.14

require (
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20200817094157-8392259f5be1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	google.golang.org/genproto v0.0.0-20190916214212-f660b8655731 // indirect
	gotest.tools/v3 v3.0.2 // indirect
	helm.sh/helm/v3 v3.3.0
	istio.io/client-go v0.0.0-20200814134724-bcbf0ed82b30
	istio.io/gogo-genproto v0.0.0-20200130224810-a0338448499a // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime v0.18.8
	k8s.io/client-go v0.18.8
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.1
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/text => golang.org/x/text v0.3.3
)
