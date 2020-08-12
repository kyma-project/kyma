module github.com/kyma-project/kyma/tests/application-operator-tests

go 1.14

require (
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/kubernetes-sigs/service-catalog v0.2.2
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20200610124020-b925f71c3209
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	helm.sh/helm/v3 v3.1.3
	istio.io/api v0.0.0-20200316215140-da46fe8e25be // indirect
	istio.io/client-go v0.0.0-20200316192452-065c59267750
	istio.io/gogo-genproto v0.0.0-20200130224810-a0338448499a // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/cli-runtime v0.17.2
	k8s.io/client-go v0.17.2
)

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
