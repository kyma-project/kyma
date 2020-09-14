module github.com/kyma-project/kyma/tests/application-gateway-tests

go 1.14

require (
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20200903161647-0fae3728c173
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.2.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
)

replace golang.org/x/text => golang.org/x/text v0.3.3
