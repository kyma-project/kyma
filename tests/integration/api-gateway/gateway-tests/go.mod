module github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests

go 1.12

require (
	github.com/avast/retry-go v2.4.1+incompatible
	github.com/kyma-project/kyma v0.5.1-0.20190909070658-69599d4a33a2
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.3.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/net v0.0.0-20190812203447-cdfb69ac37fc
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sys v0.0.0-20191010194322-b09406accb47 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	k8s.io/apimachinery v0.0.0-20191006235458-f9f2f3f8ab02
	k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
	sigs.k8s.io/controller-runtime v0.3.0 // indirect
	sigs.k8s.io/yaml v1.1.0
)
