module github.com/kyma-project/kyma/tests/application-gateway-tests

go 1.14

require (
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/kubernetes-incubator/service-catalog v0.2.2
	github.com/kubernetes-sigs/service-catalog v0.2.2
	github.com/kyma-project/kyma/common v0.0.0-20200904082145-acbc3f64c1a2
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20200825094731-2ab8b8780e41
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v0.18.3
)
