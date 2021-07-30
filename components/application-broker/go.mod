module github.com/kyma-project/kyma/components/application-broker

go 1.16

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9-0.20191108183826-59d068f8d8ff // indirect
	github.com/99designs/gqlgen v0.10.2 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/aws/aws-sdk-go v1.29.12 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/google/go-cmp v0.5.1
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.13.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/imdario/mergo v0.3.9
	github.com/kubernetes-sigs/go-open-service-broker-client v0.0.0-20200527163240-4406bd2cb6b8
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20201110114731-9af1781d40a1
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20201110134855-a03ec1689c4e
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20201110134855-a03ec1689c4e
	github.com/matryer/is v1.4.0
	github.com/mcuadros/go-defaults v1.1.0
	github.com/meatballhat/negroni-logrus v1.1.1-0.20191208165538-6f72cade44a3
	github.com/oklog/ulid v1.3.1
	github.com/openzipkin/zipkin-go v0.2.5 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/negroni v1.0.0
	github.com/vrischmann/envconfig v1.2.0
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.14.0 // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0
	gopkg.in/yaml.v2 v2.3.0
	istio.io/api v0.0.0-20201113155354-fcf32ac5d223
	istio.io/client-go v0.0.0-20201113160737-d4c1e2c0a42f
	istio.io/gogo-genproto v0.0.0-20191029161641-f7d19ec0141d // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	knative.dev/pkg v0.0.0-20200113182502-b8dc5fbc6d2f
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace (
	github.com/kubernetes-sigs/service-catalog => github.com/kubernetes-sigs/service-catalog v0.2.2-0.20190920221325-ccab52343967
	github.com/kyma-project/kyma/components/application-operator => github.com/kyma-project/kyma/components/application-operator v0.0.0-20200610105106-1066324c83cd
	google.golang.org/grpc => google.golang.org/grpc v1.27.1

	istio.io/api => istio.io/api v0.0.0-20200724154434-34e474846e0d
	istio.io/client-go => istio.io/client-go v0.0.0-20200513180646-f8d9d8ff84e6
	k8s.io/api => k8s.io/api v0.16.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.10
	k8s.io/client-go => k8s.io/client-go v0.16.10
)
