module github.com/kyma-project/kyma/components/console-backend-service

go 1.12

require (
	github.com/99designs/gqlgen v0.11.3
	github.com/blang/semver v3.5.0+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/googleapis/gnostic v0.4.0
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-incubator/api-gateway v0.0.0-20191125140217-295e8fcaa03f
	github.com/kyma-project/helm-broker v0.0.0-20190906085923-d07feb2d365a
	github.com/kyma-project/kyma v0.5.1-0.20200824115234-c78d8c28ae20
	github.com/kyma-project/kyma/common v0.0.0-20200824115234-c78d8c28ae20
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20200824115234-c78d8c28ae20
	github.com/kyma-project/kyma/components/function-controller v0.0.0-20200824091842-73a94348bb7d
	github.com/kyma-project/kyma/components/service-binding-usage-controller v0.0.0-20200824115234-c78d8c28ae20
	github.com/kyma-project/rafter v0.0.0-20200402080904-a0157e52e150
	github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.2.3
	github.com/ory/hydra-maester v0.0.19
	github.com/ory/oathkeeper-maester v0.0.7
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/vektah/gqlparser/v2 v2.0.1
	github.com/vrischmann/envconfig v1.2.0
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.14.0 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.4
	k8s.io/apiserver v0.18.4
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	knative.dev/eventing v0.13.1
	knative.dev/pkg v0.0.0-20200306230727-a56a6ea3fa56
)

replace (
	golang.org/x/text => golang.org/x/text v0.3.3
	k8s.io/client-go => k8s.io/client-go v0.18.4
)
