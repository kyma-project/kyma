module github.com/kyma-project/kyma/tests/application-connector-tests

go 1.15

require (
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20200610105106-1066324c83cd
	github.com/kyma-project/kyma/components/connection-token-handler v0.0.0-20200910095128-2407bbc2f029
	github.com/kyma-project/kyma/components/event-sources v0.0.0-20210215150315-2f01779255fe
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	istio.io/gogo-genproto v0.0.0-20200324192310-d3e214cd829a // indirect
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/pkg v0.0.0-20200625173728-dfb81cf04a7c // indirect
)

replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.15
	k8s.io/client-go => k8s.io/client-go v0.17.15
)
