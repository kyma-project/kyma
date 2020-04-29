module github.com/kyma-project/kyma/tests/function-controller

go 1.13

require (
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kyma-project/kyma/components/function-controller v0.0.0-20200429131806-d45e012bf689
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.2.0 // indirect
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200413165638-669c56c373c4 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.17.4 // indirect
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
	knative.dev/eventing v0.12.0
	knative.dev/pkg v0.0.0-20200207155214-fef852970f43
	knative.dev/serving v0.12.1
	sigs.k8s.io/controller-runtime v0.5.2
)


// mismatch among fun-controller, knative enevting and knative serving...
// try to delete it after update of eventing/serving
replace knative.dev/pkg => knative.dev/pkg v0.0.0-20200113182502-b8dc5fbc6d2f
