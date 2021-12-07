module github.com/kyma-project/kyma/components/application-broker

go 1.16

require (
	github.com/99designs/gqlgen v0.10.2 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/davecgh/go-spew v1.1.1
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/google/go-cmp v0.5.5
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/imdario/mergo v0.3.12
	github.com/kubernetes-sigs/go-open-service-broker-client v0.0.0-20200527163240-4406bd2cb6b8
	github.com/kubernetes-sigs/service-catalog v0.3.1
	github.com/kyma-incubator/compass/components/director v0.0.0-20201110114731-9af1781d40a1
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20211026115334-fd2657644868
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210624133846-3e1e71e9f682
	github.com/matryer/is v1.4.0
	github.com/mcuadros/go-defaults v1.1.0
	github.com/meatballhat/negroni-logrus v1.1.1-0.20191208165538-6f72cade44a3
	github.com/oklog/ulid v1.3.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/negroni v1.0.0
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	gopkg.in/yaml.v2 v2.4.0
	istio.io/api v0.0.0-20211012192923-310f2a3f3c76
	istio.io/client-go v1.11.4
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v0.22.0
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace (
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/kubernetes-sigs/service-catalog => github.com/kubernetes-sigs/service-catalog v0.2.2-0.20190920221325-ccab52343967
	github.com/kyma-project/kyma/components/application-operator => github.com/koala7659/kyma/components/application-operator v0.0.0-20211207101924-24d09934d9b6
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
	istio.io/api => istio.io/api v0.0.0-20200724154434-34e474846e0d
	istio.io/client-go => istio.io/client-go v0.0.0-20200513180646-f8d9d8ff84e6
	k8s.io/api => k8s.io/api v0.16.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.10
	k8s.io/client-go => k8s.io/client-go v0.16.10
)
