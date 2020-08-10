module github.com/kyma-project/kyma/components/application-operator

go 1.14

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/kubernetes-sigs/service-catalog v0.2.2
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.6.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	helm.sh/helm/v3 v3.1.3
	k8s.io/apimachinery v0.17.2
	k8s.io/cli-runtime v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/klog v1.0.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.5.0
)

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
