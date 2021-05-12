module github.com/kyma-project/kyma/components/console-backend-service

go 1.12

require (
	github.com/99designs/gqlgen v0.11.3
	github.com/blang/semver v3.5.1+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/googleapis/gnostic v0.4.0
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/kubernetes-sigs/service-catalog v0.2.2-0.20190920221325-ccab52343967
	github.com/kyma-incubator/api-gateway v0.0.0-20200930072023-5d3f2107a1ef
	github.com/kyma-project/helm-broker v0.0.0-20190906085923-d07feb2d365a
	github.com/kyma-project/kyma v0.5.1-0.20200609051543-f5997d4a36d6
	github.com/kyma-project/kyma/components/eventing-controller v0.0.0-20201127140131-ec965cad1047
	github.com/kyma-project/kyma/components/function-controller v0.0.0-20210415084126-ed2c688b52ab
	github.com/kyma-project/rafter v0.0.0-20200402080904-a0157e52e150
	github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/ory/hydra-maester v0.0.19
	github.com/ory/oathkeeper-maester v0.1.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/vektah/gqlparser/v2 v2.0.1
	github.com/vrischmann/envconfig v1.3.0
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	k8s.io/api v0.18.18
	k8s.io/apimachinery v0.18.18
	k8s.io/apiserver v0.18.12
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.13.1
	knative.dev/pkg v0.0.0-20201026165741-2f75016c1368
)

replace (
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.34.1
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible
	golang.org/x/text => golang.org/x/text v0.3.3
	k8s.io/api => k8s.io/api v0.18.18
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.18
	k8s.io/apiserver => k8s.io/apiserver v0.18.18
	k8s.io/client-go => k8s.io/client-go v0.18.18
	k8s.io/utils => k8s.io/utils v0.0.0-20191114184206-e782cd3c129f
	// kyma/components/function-controller uses new version knative.dev, with another impelmntations functions.
	knative.dev/pkg => knative.dev/pkg v0.0.0-20210217160502-b7d7ff183788
)
