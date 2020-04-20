module github.com/kyma-project/kyma/components/function-controller

go 1.13

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/go-containerregistry v0.0.0-20200331213917-3d03ed9b1ca2 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.2.0
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.17.3 // indirect
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
	knative.dev/pkg v0.0.0-20200207155214-fef852970f43
	knative.dev/serving v0.12.1
	sigs.k8s.io/controller-runtime v0.5.1
)
