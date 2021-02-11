module github.com/kyma-project/kyma/components/function-controller

go 1.15

replace (
	go.etcd.io/etcd => go.etcd.io/etcd v3.3.25+incompatible
	golang.org/x/text => golang.org/x/text v0.3.3
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.3.0
	k8s.io/client-go => k8s.io/client-go v0.18.12
)

require (
	github.com/go-git/go-git/v5 v5.1.0
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.12
	k8s.io/apimachinery v0.18.12
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19
	knative.dev/pkg v0.0.0-20201026165741-2f75016c1368
	sigs.k8s.io/controller-runtime v0.6.5
)
