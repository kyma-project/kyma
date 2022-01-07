module github.com/kyma-project/kyma/components/function-controller

go 1.16

replace (
	go.etcd.io/etcd => go.etcd.io/etcd v3.3.25+incompatible
	golang.org/x/text => golang.org/x/text v0.3.5
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.3.0
	k8s.io/client-go => k8s.io/client-go v0.18.18
)

require (
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/libgit2/git2go/v31 v31.4.14
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20211202192323-5770296d904e
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.18
	k8s.io/apimachinery v0.18.18
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19 // indirect
	knative.dev/pkg v0.0.0-20210217160502-b7d7ff183788
	sigs.k8s.io/controller-runtime v0.6.5
)
