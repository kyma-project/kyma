module github.com/dariadomagala/kyma/common

go 1.14

require (
	github.com/avast/retry-go v2.2.0+incompatible
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	google.golang.org/appengine v1.6.2 // indirect
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v0.18.3
	k8s.io/code-generator v0.0.0-00010101000000-000000000000
)

replace k8s.io/code-generator => k8s.io/code-generator v0.18.6
