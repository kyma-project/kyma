module github.com/kyma-project/kyma/components/application-registry

go 1.14

require (
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/go-openapi/spec v0.19.4
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20200812083241-407f5a1f9fab
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20200812083241-407f5a1f9fab
	github.com/kyma-project/rafter v0.0.0-20191209072228-90c07ef7c8a3
	github.com/prometheus/client_golang v1.2.0
	github.com/prometheus/client_model v0.1.0 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20191207000613-e7e4b65ae663 // indirect
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6 // indirect
	golang.org/x/sys v0.0.0-20191206220618-eeba5f6aabab // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
)

replace (
	github.com/kyma-project/kyma/components/application-gateway => github.com/franpog859/kyma/components/application-gateway v0.0.0-20200812095246-cd67ce3de73c
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.8.0
	k8s.io/client-go => k8s.io/client-go v0.17.2
)
