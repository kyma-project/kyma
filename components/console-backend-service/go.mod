module github.com/kyma-project/kyma/components/console-backend-service

go 1.12

require (
	github.com/99designs/gqlgen v0.11.3
	github.com/Shopify/sarama v1.26.0 // indirect
	github.com/apache/thrift v0.13.0 // indirect
	github.com/blang/semver v3.5.0+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.0
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/klauspost/compress v1.9.8 // indirect
	github.com/knative/eventing v0.13.1
	github.com/kubernetes-sigs/service-catalog v0.2.2-0.20190920221325-ccab52343967
	github.com/kyma-incubator/api-gateway v0.0.0-20191125140217-295e8fcaa03f
	github.com/kyma-project/helm-broker v0.0.0-20190906085923-d07feb2d365a
	github.com/kyma-project/kyma v0.5.1-0.20200609051543-f5997d4a36d6
	github.com/kyma-project/kyma/components/function-controller v0.0.0-20200527102940-1579eff8c7a5
	github.com/kyma-project/rafter v0.0.0-20200402080904-a0157e52e150
	github.com/moby/moby v1.13.1
	github.com/onsi/gomega v1.9.0
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5 // indirect
	github.com/openzipkin/zipkin-go-opentracing v0.3.5
	github.com/ory/oathkeeper-maester v0.0.7
	github.com/pierrec/lz4 v2.4.1+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.4.0
	github.com/vektah/gqlparser/v2 v2.0.1
	github.com/vrischmann/envconfig v1.2.0
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.14.0 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/jcmturner/gokrb5.v7 v7.4.0 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/apiserver v0.17.4
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	knative.dev/eventing v0.13.1 // indirect
	knative.dev/pkg v0.0.0-20200306005226-fc857aa77f79
)

replace k8s.io/client-go => k8s.io/client-go v0.17.4
