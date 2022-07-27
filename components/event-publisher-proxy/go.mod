module github.com/kyma-project/kyma/components/event-publisher-proxy

go 1.18

require (
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.10.1
	github.com/cloudevents/sdk-go/v2 v2.10.1
	github.com/google/uuid v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20220620092658-f1f8e3668674
	github.com/kyma-project/kyma/components/eventing-controller v0.0.0-20220629063704-83cc2f69ed02
	github.com/nats-io/nats-server/v2 v2.8.4
	github.com/nats-io/nats.go v1.16.0
	github.com/onsi/gomega v1.19.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.2
	github.com/stretchr/testify v1.8.0
	go.opencensus.io v0.23.0
	go.uber.org/zap v1.21.0
	golang.org/x/oauth2 v0.0.0-20220608161450-d0670ef3b1eb
	k8s.io/api v0.24.3
	k8s.io/apimachinery v0.24.3
	k8s.io/client-go v0.24.3
	sigs.k8s.io/controller-runtime v0.12.3
)

require (
	cloud.google.com/go v0.81.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.8.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.14.4 // indirect
	github.com/kyma-project/kyma/common/logging v0.0.0-20220628153305-4e32e8abea0f // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/jwt/v2 v2.2.1-0.20220330180145-442af02fd36a // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20220331220935-ae2d96664a29 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220603121420-31174f50af60 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	// keep all following packages at the same version
	k8s.io/api => k8s.io/api v0.24.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.24.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.24.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.24.3
	k8s.io/component-base => k8s.io/component-base v0.24.3
	k8s.io/kubectl => k8s.io/kubectl v0.24.3
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1 // fix CVE-2020-26160
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/docker/docker v20.10.3+incompatible
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc93
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.12.2
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20220331220935-ae2d96664a29
	k8s.io/utils => k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
)

replace (
	github.com/emicklei/go-restful => github.com/emicklei/go-restful/v3 v3.8.0
	github.com/emicklei/go-restful/v3 => github.com/emicklei/go-restful/v3 v3.8.0
) // fix cve-2022-1996, might be okay to be removed after k8s > 1.24.2

exclude (
	github.com/emicklei/go-restful v0.0.0-20170410110728-ff4f55a20633
	github.com/emicklei/go-restful v2.9.5+incompatible
) // fix cve-2022-1996, might be okay to be removed after k8s > 1.24.2
