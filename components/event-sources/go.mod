module github.com/kyma-project/kyma/components/event-sources

go 1.15

require (
	cloud.google.com/go v0.49.0 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9-0.20191108183826-59d068f8d8ff // indirect
	contrib.go.opencensus.io/exporter/zipkin v0.1.1 // indirect
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/aws/aws-sdk-go v1.25.45 // indirect
	github.com/cloudevents/sdk-go v0.11.0
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.4 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/client_model v0.1.0 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	go.opencensus.io v0.22.2
	go.uber.org/atomic v1.5.1 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.13.0
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/genproto v0.0.0-20191203145615-049a07e0debe // indirect

	// istio version "1.5.8"
	istio.io/api v0.0.0-20200513175333-ae3da0d240e3
	istio.io/client-go v0.0.0-20200513180646-f8d9d8ff84e6
	istio.io/gogo-genproto v0.0.0-20191029161641-f7d19ec0141d // indirect
	k8s.io/api v0.18.1
	k8s.io/apiextensions-apiserver v0.16.16-rc.0 // indirect
	k8s.io/apimachinery v0.18.1
	k8s.io/client-go v0.17.15
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
	knative.dev/eventing v0.12.0
	knative.dev/pkg v0.0.0-20200113182502-b8dc5fbc6d2f
)

replace (
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2 // fix CVE-2021-3121
	k8s.io/api => k8s.io/api v0.17.15
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.15
	k8s.io/client-go => k8s.io/client-go v0.17.15
	knative.dev/pkg => knative.dev/pkg v0.0.0-20200113182502-b8dc5fbc6d2f
)
