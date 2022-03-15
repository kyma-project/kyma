module github.com/kyma-project/kyma/components/function-controller

go 1.16

replace (
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.42.52
	github.com/prometheus/common => github.com/prometheus/common v0.26.0
	go.etcd.io/etcd => go.etcd.io/etcd v3.3.25+incompatible
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220213190939-1e6e3497d506
	golang.org/x/net => golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/text => golang.org/x/text v0.3.8-0.20220124021120-d1c84af989ab
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.3.0
	k8s.io/api => k8s.io/api v0.22.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.7
	k8s.io/client-go => k8s.io/client-go v0.22.7
	k8s.io/component-base => k8s.io/component-base v0.22.7
)

require (
	github.com/go-logr/logr v1.2.2
	github.com/go-logr/zapr v1.2.0
	github.com/libgit2/git2go/v31 v31.4.14
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
	knative.dev/pkg v0.0.0-20220310182448-ce6f2877cef9
	sigs.k8s.io/controller-runtime v0.11.1
)
