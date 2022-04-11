module github.com/kyma-project/kyma/components/application-broker

go 1.18

require (
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/davecgh/go-spew v1.1.1
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/google/go-cmp v0.5.5
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/imdario/mergo v0.3.9
	github.com/kubernetes-sigs/go-open-service-broker-client v0.0.0-20200527163240-4406bd2cb6b8
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20201110114731-9af1781d40a1
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20201110134855-a03ec1689c4e
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20201110134855-a03ec1689c4e
	github.com/matryer/is v1.4.0
	github.com/mcuadros/go-defaults v1.2.0
	github.com/meatballhat/negroni-logrus v0.0.0-20201129033903-bc51654b0848
	github.com/oklog/ulid v1.3.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.6.1
	github.com/urfave/negroni v1.0.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	gopkg.in/yaml.v2 v2.4.0
	istio.io/api v0.0.0-20201113155354-fcf32ac5d223
	istio.io/client-go v0.0.0-20201113160737-d4c1e2c0a42f
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
)

require (
	github.com/99designs/gqlgen v0.10.2 // indirect
	github.com/agnivade/levenshtein v1.0.3 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/vektah/gqlparser v1.3.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.1.1-0.20190912152152-6a016cf16650 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c // indirect
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/genproto v0.0.0-20200825200019-8632dd797987 // indirect
	google.golang.org/grpc v1.31.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
	istio.io/gogo-genproto v0.0.0-20191029161641-f7d19ec0141d // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200410163147-594e756bea31 // indirect
	k8s.io/utils v0.0.0-20191114184206-e782cd3c129f // indirect
	sigs.k8s.io/controller-runtime v0.5.0 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace (
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/kubernetes-sigs/service-catalog => github.com/kubernetes-sigs/service-catalog v0.2.2-0.20190920221325-ccab52343967
	github.com/kyma-project/kyma/components/application-operator => github.com/kyma-project/kyma/components/application-operator v0.0.0-20200610105106-1066324c83cd
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
	istio.io/api => istio.io/api v0.0.0-20200724154434-34e474846e0d
	istio.io/client-go => istio.io/client-go v0.0.0-20200513180646-f8d9d8ff84e6
	k8s.io/api => k8s.io/api v0.16.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.10
	k8s.io/client-go => k8s.io/client-go v0.16.10
)
