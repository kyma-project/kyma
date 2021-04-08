module github.com/kyma-project/kyma/components/event-publisher-proxy

go 1.15

require (
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.3.1
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kyma-project/kyma v0.5.1-0.20200609051543-f5997d4a36d6
	github.com/kyma-project/kyma/components/eventing-controller v0.0.0-20210112215829-419ae45b5316
	github.com/nats-io/nats-server/v2 v2.1.9
	github.com/nats-io/nats.go v1.10.1-0.20201204000952-090c71e95cd0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	go.opencensus.io v0.22.4
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.14.0 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	golang.org/x/tools v0.0.0-20200921210052-fa0125251cc4 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/api v0.19.8
	k8s.io/apimachinery v0.19.8
	k8s.io/client-go v0.19.8
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/controller-runtime v0.6.0
)

replace github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1 // fix CVE-2020-26160
