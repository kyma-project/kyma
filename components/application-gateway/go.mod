module github.com/kyma-project/kyma/components/application-gateway

go 1.14

require (
	github.com/gorilla/mux v1.7.4
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20200810104940-0669e658a4a6
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
)

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
