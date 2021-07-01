module github.com/kyma-project/kyma/components/connection-token-handler

go 1.16

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/go-logr/zapr v0.4.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.17.0
	golang.org/x/tools v0.1.4 // indirect
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/code-generator v0.21.2
	sigs.k8s.io/controller-runtime v0.9.2

)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/text => golang.org/x/text v0.3.3
)
