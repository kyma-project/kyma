module github.com/kyma-project/kyma/components/application-registry

go 1.14

require (
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/go-openapi/spec v0.19.4
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20200825094731-2ab8b8780e41
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20200902071617-83c683010f30
	github.com/kyma-project/rafter v0.0.0-20191209072228-90c07ef7c8a3
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6 // indirect
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.4
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

replace (
	github.com/asaskevich/govalidator => github.com/asaskevich/govalidator v0.0.0-20180315120708-ccb8e960c48f
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.8.0
	k8s.io/client-go => k8s.io/client-go v0.17.2
)
