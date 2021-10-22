module github.com/kyma-project/kyma/components/eventing-controller

go 1.16

require (
	github.com/avast/retry-go/v3 v3.1.1
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.5.0
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kyma-incubator/api-gateway v0.0.0-20210909101151-ac45e9ce3553
	github.com/kyma-project/kyma/common/logging v0.0.0-20211006112227-6d16a34ea468
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20211006112227-6d16a34ea468
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/nats-io/nats-server/v2 v2.6.1
	github.com/nats-io/nats.go v1.12.3
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.18.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v0.22.0
	sigs.k8s.io/controller-runtime v0.9.6
)
