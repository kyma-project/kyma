module github.com/kyma-project/kyma/tests/application-gateway-tests

go 1.14

require (
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/kubernetes-incubator/service-catalog v0.2.2
	github.com/kubernetes-sigs/service-catalog v0.2.2
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20200812083241-407f5a1f9fab
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
)

replace github.com/kyma-project/kyma/components/application-gateway => github.com/franpog859/kyma/components/application-gateway v0.0.0-20200812095246-cd67ce3de73c
