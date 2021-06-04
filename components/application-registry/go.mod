module github.com/kyma-project/kyma/components/application-registry

go 1.16

require (
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/go-openapi/spec v0.19.4
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/kyma-project/kyma/components/application-gateway v0.0.0-20210604090801-9a75cdfd1c7c
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210604090801-9a75cdfd1c7c
	github.com/kyma-project/rafter v0.0.0-20210202141112-0bd2218c9c12
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/code-generator v0.20.2
)

replace (
	github.com/asaskevich/govalidator => github.com/asaskevich/govalidator v0.0.0-20180315120708-ccb8e960c48f
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/docker/docker v20.10.3+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc92

	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/text => golang.org/x/text v0.3.3

)
