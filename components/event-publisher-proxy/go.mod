module github.com/kyma-project/kyma/components/event-publisher-proxy

go 1.15

require (
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.3.1
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/google/uuid v1.1.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210520105418-ddc3a476c40a
	github.com/kyma-project/kyma/components/eventing-controller v0.0.0-20210112215829-419ae45b5316
	github.com/nats-io/nats-server/v2 v2.2.4
	github.com/nats-io/nats.go v1.11.0
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.22.4
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/tools v0.1.1 // indirect
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	sigs.k8s.io/controller-runtime v0.8.3
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1 // fix CVE-2020-26160
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/docker/docker v20.10.3+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc93
)
