module github.com/kyma-project/kyma/common

go 1.18

require (
	github.com/avast/retry-go v2.2.0+incompatible
	github.com/stretchr/testify v1.7.0
	golang.org/x/tools v0.1.12
	k8s.io/code-generator v0.18.6
)

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/go-openapi/jsonpointer v0.19.3 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.0.0-20220728030405-41545e8bf201 // indirect
	golang.org/x/sys v0.0.0-20220728004956-3c1f35247d10 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
	k8s.io/gengo v0.0.0-20200114144118-36b2048a9120 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
	golang.org/x/net => golang.org/x/net v0.0.0-20220728030405-41545e8bf201
	golang.org/x/text => golang.org/x/text v0.3.7
	golang.org/x/tools => golang.org/x/tools v0.1.12
)
