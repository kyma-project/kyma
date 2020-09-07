module github.com/kyma-project/kyma/tests/function-controller

go 1.15

require (
	github.com/Azure/go-autorest/autorest v0.9.3 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.1 // indirect
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/go-git/go-git/v5 v5.1.0
	github.com/hashicorp/errwrap v1.0.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/kubernetes-incubator/service-catalog v0.2.2
	github.com/kubernetes-sigs/service-catalog v0.2.2 // indirect
	github.com/kyma-project/helm-broker v1.0.0
	github.com/kyma-project/kyma/components/function-controller v0.0.0-20200902071617-83c683010f30
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/pflag v1.0.5
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/multierr v1.5.0 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/sys v0.0.0-20200810151505-1b9f1253b3ed // indirect
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
	k8s.io/api v0.17.11
	k8s.io/apimachinery v0.17.11
	k8s.io/client-go v0.17.11
	knative.dev/eventing v0.12.0
	knative.dev/pkg v0.0.0-20200306230727-a56a6ea3fa56
	sigs.k8s.io/controller-runtime v0.5.10
)

replace (
	go.etcd.io/etcd => go.etcd.io/etcd v3.3.25+incompatible

	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a

	golang.org/x/text => golang.org/x/text v0.3.3

	// mismatch among fun-controller, knative enevting and knative serving...
	// try to delete it after update of eventing/serving
	knative.dev/pkg => knative.dev/pkg v0.0.0-20200113182502-b8dc5fbc6d2f
)
