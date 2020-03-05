package container

import (
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/pkg/resource-watcher"
	"k8s.io/client-go/dynamic"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Container struct {
	Manager                 ctrl.Manager
	CoreClient              *v1.CoreV1Client
	DynamicClient           *dynamic.Interface
	ResourceWatcherServices *resource_watcher.Services
}
