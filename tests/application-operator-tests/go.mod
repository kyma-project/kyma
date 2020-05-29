module github.com/kyma-project/kyma/tests/application-operator-tests

go 1.13

require (
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.19.0+incompatible // indirect
	github.com/kubernetes-sigs/service-catalog v0.2.2
	github.com/kyma-project/kyma v0.5.1-0.20190822062921-70782d188811
	github.com/stretchr/testify v1.4.0
	helm.sh/helm/v3 v3.1.3
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
	istio.io/api v0.0.0-20200316215140-da46fe8e25be // indirect
	istio.io/client-go v0.0.0-20200316192452-065c59267750
	istio.io/gogo-genproto v0.0.0-20200130224810-a0338448499a // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/cli-runtime v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/helm v2.16.7+incompatible
	k8s.io/klog v1.0.0
)

replace github.com/kyma-project/kyma => github.com/kyma-project/kyma v0.5.1-0.20190822062921-70782d188811
