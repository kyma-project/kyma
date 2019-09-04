module github.com/kyma-project/kyma/components/asset-store-controller-manager

go 1.12

require (
	github.com/go-ini/ini v1.46.0 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/minio/minio-go v6.0.14+incompatible
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/onsi/ginkgo v1.6.0
	github.com/onsi/gomega v1.4.2
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.3.0
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0
)
