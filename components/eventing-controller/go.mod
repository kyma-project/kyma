module github.com/kyma-project/kyma/components/eventing-controller

go 1.14

require (
	github.com/avast/retry-go v2.4.1+incompatible
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/go-logr/logr v0.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kyma-incubator/api-gateway v0.0.0-20200930072023-5d3f2107a1ef
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210204131215-a368a90f2525
	github.com/kyma-project/kyma/components/event-publisher-proxy v0.0.0-20201014135541-82b304ab245a
	github.com/mitchellh/hashstructure v1.0.0
	github.com/nats-io/nats-server/v2 v2.1.9
	github.com/nats-io/nats.go v1.10.1-0.20201204000952-090c71e95cd0
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/ory/oathkeeper-maester v0.1.0
	github.com/pkg/errors v0.9.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	k8s.io/api v0.19.3
	k8s.io/apiextensions-apiserver v0.19.3 // indirect
	k8s.io/apimachinery v0.19.3
	k8s.io/client-go v0.19.3
	sigs.k8s.io/controller-runtime v0.6.2
)

replace github.com/nats-io/nats.go => github.com/nats-io/nats.go v1.10.1-0.20201204000952-090c71e95cd0

replace github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2

replace github.com/go-logr/zapr => github.com/go-logr/zapr v0.3.0
