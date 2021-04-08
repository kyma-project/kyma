module github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests

go 1.12

require (
	cloud.google.com/go v0.54.0 // indirect
	github.com/Azure/go-autorest/autorest v0.11.1 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.5 // indirect
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/kyma-project/kyma/common v0.0.0-20210408154703-8f99ba66b4e3
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20201112073958-5cba982894dd // indirect
	golang.org/x/text v0.3.4 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apimachinery v0.18.16
	k8s.io/client-go v0.18.16
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.17
	k8s.io/client-go => k8s.io/client-go v0.17.17
)
