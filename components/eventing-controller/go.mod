module github.com/kyma-project/kyma/components/eventing-controller

go 1.14

require (
	github.com/avast/retry-go v2.4.1+incompatible
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.3.1
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/gobuffalo/flect v0.2.3 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kyma-incubator/api-gateway v0.0.0-20200930072023-5d3f2107a1ef
	github.com/kyma-project/kyma/common/logging v0.0.0-20210601142757-445a3b6021fe
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210204131215-a368a90f2525
	github.com/lightstep/tracecontext.go v0.0.0-20181129014701-1757c391b1ac // indirect
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/nats-io/nats-server/v2 v2.2.4
	github.com/nats-io/nats.go v1.11.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/ory/hydra-maester v0.0.22 // indirect
	github.com/ory/oathkeeper-maester v0.1.0
	github.com/pkg/errors v0.9.1
	github.com/smallstep/logging v0.1.0 // indirect
	github.com/spf13/cobra v1.2.1 // indirect
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.0
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602
	golang.org/x/tools v0.1.5 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.22.2 // indirect
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/controller-tools v0.7.0 // indirect
)

replace (
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.3.0
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/nats-io/nats.go => github.com/nats-io/nats.go v1.11.0
	k8s.io/api => k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.7
	k8s.io/client-go => k8s.io/client-go v0.20.7
)
