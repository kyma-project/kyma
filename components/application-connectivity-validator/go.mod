module github.com/kyma-project/kyma/components/application-connectivity-validator

go 1.15

require (
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/google/ko v0.6.2 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect	github.com/gorilla/mux v1.7.4
	github.com/gorilla/mux v1.7.4
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20200902071617-83c683010f30
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.13.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	k8s.io/klog/v2 v2.2.0
	sigs.k8s.io/controller-runtime v0.6.2
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.1
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc92
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/text => golang.org/x/text v0.3.3
	k8s.io/api => k8s.io/api v0.0.0-20201020200614-54bcc311e327
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20201020200440-554eef9dbf66
	k8s.io/client-go => k8s.io/client-go v0.0.0-20201020200834-d1a4fe5f2d96
)
