module github.com/kyma-project/kyma/components/application-broker

go 1.16

require (
	github.com/99designs/gqlgen v0.10.2 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/davecgh/go-spew v1.1.1
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.9
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kubernetes-sigs/go-open-service-broker-client v0.0.0-20200527163240-4406bd2cb6b8
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20201110114731-9af1781d40a1
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20201110134855-a03ec1689c4e
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20201110134855-a03ec1689c4e
	github.com/matryer/is v1.4.0
	github.com/mcuadros/go-defaults v1.1.0
	github.com/meatballhat/negroni-logrus v1.1.1-0.20191208165538-6f72cade44a3
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/oklog/ulid v1.3.1
	github.com/onsi/gomega v1.10.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.7.0
	github.com/urfave/negroni v1.0.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a // indirect
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985 // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914 // indirect
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210416161957-9910b6c460de // indirect
	google.golang.org/grpc v1.39.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	istio.io/api v0.0.0-20201113155354-fcf32ac5d223
	istio.io/client-go v0.0.0-20201113160737-d4c1e2c0a42f
	istio.io/gogo-genproto v0.0.0-20191029161641-f7d19ec0141d // indirect
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace (
	github.com/kubernetes-sigs/service-catalog => github.com/kubernetes-sigs/service-catalog v0.2.2-0.20190920221325-ccab52343967
	github.com/kyma-project/kyma/components/application-operator => github.com/kyma-project/kyma/components/application-operator v0.0.0-20200610105106-1066324c83cd
	google.golang.org/grpc => google.golang.org/grpc v1.27.1

	istio.io/api => istio.io/api v0.0.0-20200724154434-34e474846e0d
	istio.io/client-go => istio.io/client-go v0.0.0-20200513180646-f8d9d8ff84e6
)
