module github.com/kyma-project/kyma/tests/application-gateway-legacy-tests

go 1.15

require (
	github.com/avast/retry-go v2.4.1+incompatible
	github.com/gorilla/mux v1.7.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kyma-project/kyma/common v0.0.0-20201127133638-b7a5c2ad97a7
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/text v0.3.3 // indirect
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v0.18.3
)

replace (
	k8s.io/api => k8s.io/api v0.17.14
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.14
	k8s.io/client-go => k8s.io/client-go v0.17.14
)
