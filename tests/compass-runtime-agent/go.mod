module github.com/kyma-project/kyma/tests/compass-runtime-agent

go 1.14

require (
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/kubernetes-incubator/service-catalog v0.2.2
	github.com/kubernetes-sigs/service-catalog v0.3.0
	github.com/kyma-incubator/compass v0.0.0-20200608084054-64f737ad7e1d
	github.com/kyma-incubator/compass/components/director v0.0.0-20201110114731-9af1781d40a1
	github.com/kyma-project/kyma/components/application-broker v0.0.0-20210518142150-d708b6214512
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20201110134855-a03ec1689c4e
	github.com/kyma-project/kyma/tests/application-gateway-tests v0.0.0-20210518142150-d708b6214512
	github.com/kyma-project/rafter v0.0.0-20200413150919-1a89277ac3d8
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
)

replace (
	github.com/kyma-project/kyma/components/application-operator => github.com/kyma-project/kyma/components/application-operator v0.0.0-20200610105106-1066324c83cd

	k8s.io/api => k8s.io/api v0.16.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.10
	k8s.io/client-go => k8s.io/client-go v0.16.10
)
