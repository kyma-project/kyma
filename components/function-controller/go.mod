module github.com/kyma-project/kyma/components/function-controller

go 1.16

replace (
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.42.52
	github.com/prometheus/common => github.com/prometheus/common v0.26.0
	go.etcd.io/etcd => go.etcd.io/etcd v3.3.25+incompatible
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220213190939-1e6e3497d506
	golang.org/x/net => golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/text => golang.org/x/text v0.3.8-0.20220124021120-d1c84af989ab
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.3.0
	k8s.io/api => k8s.io/api v0.22.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.7
	k8s.io/client-go => k8s.io/client-go v0.22.7
	k8s.io/component-base => k8s.io/component-base v0.22.7
)

require (
	cloud.google.com/go v0.98.0 // indirect
	cloud.google.com/go/storage v1.18.2 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.7.1-0.20200907061046-05415f1de66d // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.4.0 // indirect
	contrib.go.opencensus.io/exporter/zipkin v0.1.2 // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/appscode/jsonpatch v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/coreos/go-oidc v2.1.0+incompatible // indirect
	github.com/dgryski/go-gk v0.0.0-20200319235926-a69029f61654 // indirect
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-logr/logr v1.2.2
	github.com/go-logr/zapr v1.2.0
	github.com/go-openapi/spec v0.17.2 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/gobuffalo/flect v0.2.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-github/v27 v27.0.6 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/mako v0.0.0-20190821191249-122f8dcef9e3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/libgit2/git2go/v31 v31.4.14
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.17.0
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.1
	github.com/prometheus/common v0.32.1 // indirect
	github.com/rabbitmq/amqp091-go v1.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/tsenart/vegeta/v12 v12.8.4 // indirect
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/vrischmann/envconfig v1.3.0
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg/scram v1.0.3 // indirect
	github.com/xdg/stringprep v1.0.3 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.etcd.io/etcd/client/v2 v2.305.0 // indirect
	go.etcd.io/etcd/client/v3 v3.5.0 // indirect
	go.etcd.io/etcd/pkg/v3 v3.5.0 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/automaxprocs v1.4.0 // indirect
	go.uber.org/goleak v1.1.12 // indirect
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	golang.org/x/tools v0.1.8 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/grpc v1.42.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/square/go-jose.v2 v2.2.2 // indirect
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/code-generator v0.22.7 // indirect
	k8s.io/component-base v0.23.0 // indirect
	k8s.io/gengo v0.0.0-20211129171323-c02415ce4185 // indirect
	k8s.io/klog v0.3.1 // indirect
	k8s.io/klog/v2 v2.40.1 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	knative.dev/hack v0.0.0-20220224013837-e1785985d364 // indirect
	knative.dev/pkg v0.0.0-20220302134643-d2cdc682d974
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.27 // indirect
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff v0.0.0-20190302045857-e85c7b244fd2 // indirect
	sigs.k8s.io/testing_frameworks v0.1.2 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
