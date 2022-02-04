module github.com/kyma-project/kyma/components/service-binding-usage-controller

go 1.16

require (
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.1.0
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.10
	k8s.io/client-go v0.18.4
)

replace (
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	sigs.k8s.io/structured-merge-diff/v3 => sigs.k8s.io/structured-merge-diff/v3 v3.0.0
)
