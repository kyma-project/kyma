module github.com/kyma-project/kyma/tests/application-gateway-tests

go 1.14

require (
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20200825094731-2ab8b8780e41
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.2.0
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.4
	k8s.io/client-go v0.18.4
)

replace github.com/kyma-project/kyma/components/application-gateway => github.com/Szymongib/kyma/components/application-gateway v0.0.0-20200902080656-573e032dd983
