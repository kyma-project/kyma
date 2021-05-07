module github.com/kyma-project/kyma/components/service-binding-usage-controller

go 1.16

require (
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/sirupsen/logrus v1.5.0
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.1.0
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.10
	k8s.io/client-go v0.18.4
)

replace (
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	sigs.k8s.io/structured-merge-diff/v3 => sigs.k8s.io/structured-merge-diff/v3 v3.0.0
)
