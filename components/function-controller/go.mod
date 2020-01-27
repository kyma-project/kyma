module github.com/kyma-project/kyma/components/function-controller

go 1.13

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.8 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/google/go-containerregistry v0.0.0-20200115214256-379933c9c22b // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/onsi/gomega v1.8.0
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/tektoncd/pipeline v0.7.0
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	gomodules.xyz/jsonpatch/v2 v2.0.1
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	knative.dev/pkg v0.0.0-20191230183737-ead56ad1f3bd
	knative.dev/serving v0.8.1
	sigs.k8s.io/controller-runtime v0.4.0
)
