module github.com/kyma-project/kyma/components/apiserver-proxy

go 1.14

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/handlers v1.5.1
	github.com/hkwi/h2c v0.0.0-20180807060133-3511cd63f456
	github.com/microcosm-cc/bluemonday v1.0.4
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.19.7
	k8s.io/apiserver v0.19.7
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.19.7
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.5.0 // indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.19.7
