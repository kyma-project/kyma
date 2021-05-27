module github.com/kyma-project/kyma/components/central-application-connectivity-validator

go 1.16

require (
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-cmp v0.4.1 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect; indirect	github.com/gorilla/mux v1.7.4
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kyma-project/kyma/common/logging v0.0.0-20210318081026-665ca4cda3f6
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20200902071617-83c683010f30
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.1.0
	go.uber.org/zap v1.16.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	golang.org/x/tools v0.0.0-20200924205911-8a9a89368bd3 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
	k8s.io/api v0.18.8 // indirect
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc92

	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/text => golang.org/x/text v0.3.3

	k8s.io/api => k8s.io/api v0.0.0-20201020200614-54bcc311e327
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20201020200440-554eef9dbf66
	k8s.io/client-go => k8s.io/client-go v0.0.0-20201020200834-d1a4fe5f2d96
)
