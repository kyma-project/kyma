module github.com/kyma-project/kyma/components/console-backend-service

go 1.12

require (
	github.com/99designs/gqlgen v0.11.3
	github.com/Shopify/sarama v1.26.0 // indirect
	github.com/blang/semver v3.5.0+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/googleapis/gnostic v0.4.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/klauspost/compress v1.9.8 // indirect
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-incubator/api-gateway v0.0.0-20200930072023-5d3f2107a1ef
	github.com/kyma-project/helm-broker v0.0.0-20190906085923-d07feb2d365a
	github.com/kyma-project/kyma/common v0.0.0-20210408140744-33d08542b233
	github.com/kyma-project/kyma/components/application-broker v0.0.0-20210408140744-33d08542b233
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210408140744-33d08542b233
	github.com/kyma-project/kyma/components/eventing-controller v0.0.0-20201127140131-ec965cad1047
	github.com/kyma-project/kyma/components/function-controller v0.0.0-20201012150043-858bc2c23ef5
	github.com/kyma-project/kyma/components/service-binding-usage-controller v0.0.0-20210408154703-8f99ba66b4e3
	github.com/kyma-project/rafter v0.0.0-20200402080904-a0157e52e150
	github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go-opentracing v0.3.5
	github.com/ory/hydra-maester v0.0.19
	github.com/ory/oathkeeper-maester v0.1.0
	github.com/pierrec/lz4 v2.4.1+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/vektah/gqlparser/v2 v2.0.1
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/tools v0.0.0-20200921210052-fa0125251cc4 // indirect
	gopkg.in/jcmturner/gokrb5.v7 v7.4.0 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	k8s.io/api v0.18.15
	k8s.io/apimachinery v0.18.16
	k8s.io/apiserver v0.18.8
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	knative.dev/eventing v0.13.1
	knative.dev/pkg v0.0.0-20200306230727-a56a6ea3fa56
)

replace (
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.34.1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible
	github.com/kubernetes-sigs/service-catalog => github.com/kubernetes-sigs/service-catalog v0.2.2
	github.com/openzipkin-contrib/zipkin-go-opentracing => github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5
	golang.org/x/text => golang.org/x/text v0.3.3
	k8s.io/api => k8s.io/api v0.17.11
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.11
	k8s.io/apiserver => k8s.io/apiserver v0.17.9
	k8s.io/client-go => k8s.io/client-go v0.17.4
	k8s.io/utils => k8s.io/utils v0.0.0-20191114184206-e782cd3c129f
)
