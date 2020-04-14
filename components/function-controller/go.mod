module github.com/kyma-project/kyma/components/function-controller

go 1.13

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tektoncd/pipeline v0.10.1
	k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver v0.17.3 // indirect
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	knative.dev/pkg v0.0.0-20191230183737-ead56ad1f3bd
	knative.dev/serving v0.12.1
	sigs.k8s.io/controller-runtime v0.5.1
)
