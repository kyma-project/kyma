module github.com/kyma-project/kyma/components/function-controller

go 1.13

require (
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/onsi/gomega v1.8.0
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/tektoncd/pipeline v0.10.0
	golang.org/x/net v0.0.0-20191119073136-fc4aabc6c914
	gomodules.xyz/jsonpatch/v2 v2.0.1
	k8s.io/api v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v0.17.1
	knative.dev/pkg v0.0.0-20191230183737-ead56ad1f3bd
	knative.dev/serving v0.8.1
	sigs.k8s.io/controller-runtime v0.4.0
)
