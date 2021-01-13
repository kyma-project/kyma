module github.com/kyma-project/kyma/components/event-publisher-proxy

go 1.15

require (
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.3.1
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/go-logr/logr v0.3.0
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kyma-project/kyma/components/console-backend-service v0.0.0-20201116133707-dd0a4cf8e9d8 // indirect
	github.com/kyma-project/kyma/components/eventing-controller v0.0.0-20210112215829-419ae45b5316
	github.com/magiconair/properties v1.8.0 // indirect
	github.com/nats-io/nats.go v1.10.1-0.20201204000952-090c71e95cd0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	go.opencensus.io v0.22.4
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	k8s.io/api => k8s.io/api v0.16.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.15
	k8s.io/client-go => k8s.io/client-go v0.16.15
)
