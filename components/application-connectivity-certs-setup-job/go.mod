module github.com/kyma-project/kyma/components/application-connectivity-certs-setup-job

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	google.golang.org/appengine v1.6.6 // indirect
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v0.22.0
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20211115234514-b4de73f9ece8
	golang.org/x/text => golang.org/x/text v0.3.3
)
