module github.com/kyma-project/kyma/tests/function-controller

go 1.14

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.13.1 // indirect
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/google/go-containerregistry v0.0.0-20200430153450-5cbd060f5c92 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kubernetes-incubator/service-catalog v0.2.2
	github.com/kubernetes-sigs/service-catalog v0.2.2 // indirect
	github.com/kyma-project/helm-broker v1.0.0
	github.com/kyma-project/kyma/components/function-controller v0.0.0-20200526093810-45c748b13083
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200413165638-669c56c373c4 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.17.4 // indirect
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
	knative.dev/eventing v0.14.0
	knative.dev/pkg v0.0.0-20200513151758-7b6bb61326ae
	knative.dev/serving v0.14.0
	sigs.k8s.io/controller-runtime v0.5.2
)

replace knative.dev/serving => knative.dev/serving v0.14.0
