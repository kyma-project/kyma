module github.com/kyma-project/kyma/tests/connection-token-handler-tests

go 1.14

require (
	github.com/kyma-project/kyma/components/connection-token-handler v0.0.0-20200812083241-407f5a1f9fab
	github.com/stretchr/testify v1.3.0
	k8s.io/apimachinery v0.16.11-rc.0
	k8s.io/client-go v0.16.10
)

replace github.com/kyma-project/kyma/components/connection-token-handler => github.com/franpog859/kyma/components/connection-token-handler v0.0.0-20200812095246-cd67ce3de73c
