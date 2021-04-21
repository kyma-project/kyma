module github.com/kyma-project/kyma/components/application-connectivity-certs-setup-job

go 1.16

require (
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.6.0
	golang.org/x/crypto v0.0.0-20200602180216-279210d13fed // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200602225109-6fdc65e7d980 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200603094226-e3079894b1e8 // indirect
	k8s.io/api v0.18.15
	k8s.io/apimachinery v0.18.15
	k8s.io/client-go v0.18.15
	k8s.io/kube-openapi v0.0.0-20200410163147-594e756bea31 // indirect
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451 // indirect
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/text => golang.org/x/text v0.3.3
)
