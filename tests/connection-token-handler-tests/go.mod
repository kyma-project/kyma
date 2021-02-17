module github.com/kyma-project/kyma/tests/connection-token-handler-tests

go 1.14

require (
	github.com/kyma-project/kyma/components/connection-token-handler v0.0.0-20200825094731-2ab8b8780e41
	github.com/stretchr/testify v1.3.0
	k8s.io/apimachinery v0.16.11-rc.0
	k8s.io/client-go v0.16.10
)

replace golang.org/x/text => golang.org/x/text v0.3.3
