module github.com/kyma-project/kyma/components/event-sources

go 1.15

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9-0.20191108183826-59d068f8d8ff // indirect
	contrib.go.opencensus.io/exporter/zipkin v0.1.1 // indirect
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/aws/aws-sdk-go v1.29.12 // indirect
	github.com/cloudevents/sdk-go v0.11.0
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/google/go-cmp v0.5.2
	github.com/grpc-ecosystem/grpc-gateway v1.14.4 // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210204131215-a368a90f2525
	github.com/kyma-project/kyma/components/eventing-controller v0.0.0-20210304101551-87a6c72905da
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2
	github.com/openzipkin/zipkin-go v0.2.5 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/prometheus/procfs v0.2.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	github.com/sirupsen/logrus v1.6.0
	go.opencensus.io v0.22.4
	go.uber.org/zap v1.15.0
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect

	// istio version "1.5.8"
	istio.io/api v0.0.0-20200812202721-24be265d41c3
	istio.io/client-go v0.0.0-20200916161914-94f0e83444ca
	istio.io/gogo-genproto v0.0.0-20200513175746-ea37dac65660 // indirect
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.1 // indirect
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog/v2 v2.4.0 // indirect
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
	knative.dev/eventing v0.12.0
	knative.dev/pkg v0.0.0-20200113182502-b8dc5fbc6d2f
	sigs.k8s.io/controller-runtime v0.6.5
)

replace (
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2 // fix CVE-2021-3121
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	istio.io/api => istio.io/api v0.0.0-20200513175333-ae3da0d240e3
	istio.io/client-go => istio.io/client-go v0.0.0-20200513180646-f8d9d8ff84e6
	istio.io/gogo-genproto => istio.io/gogo-genproto v0.0.0-20200513175746-ea37dac65660
	k8s.io/api => k8s.io/api v0.17.15
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.15
	k8s.io/client-go => k8s.io/client-go v0.17.15
	knative.dev/pkg => knative.dev/pkg v0.0.0-20200113182502-b8dc5fbc6d2f
)
