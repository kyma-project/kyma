module github.com/kyma-project/kyma/components/connection-token-handler

go 1.16

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.6.0 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/stretchr/testify v1.4.0
	k8s.io/apimachinery v0.18.15
	k8s.io/client-go v0.18.15
	k8s.io/code-generator v0.18.15
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/text => golang.org/x/text v0.3.3
	k8s.io/api => k8s.io/api v0.18.15
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.15
	k8s.io/apiserver => k8s.io/apiserver v0.18.15
	k8s.io/client-go => k8s.io/client-go v0.18.15
	k8s.io/code-generator => k8s.io/code-generator v0.18.15
	k8s.io/component-base => k8s.io/component-base v0.18.15
)
