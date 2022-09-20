module github.com/kyma-project/kyma/tests/function-controller

go 1.15

require (
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/go-git/go-git/v5 v5.2.0
	github.com/hashicorp/errwrap v1.0.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/kyma-project/kyma/components/eventing-controller v0.0.0-20220915084356-1d9b39c6797a
	github.com/kyma-project/kyma/components/function-controller v0.0.0-20210708083136-5479837a0948
	github.com/onsi/gomega v1.20.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
	k8s.io/api v0.25.0
	k8s.io/apimachinery v0.25.0
	k8s.io/client-go v0.25.0
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1
	//TODO: replace it in another PR
	github.com/kyma-project/kyma/components/function-controller => github.com/pPrecel/kyma/components/function-controller v0.0.0-20220801083447-41d01d8dc0c7
	go.etcd.io/etcd => go.etcd.io/etcd v3.3.25+incompatible
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/text => golang.org/x/text v0.3.3
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
)

replace k8s.io/api => k8s.io/api v0.18.12

replace k8s.io/apimachinery => k8s.io/apimachinery v0.18.12

replace k8s.io/client-go => k8s.io/client-go v0.18.12

replace k8s.io/component-base => k8s.io/component-base v0.18.12
