package ui

import (
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/resource"
)

func NewBackendModuleEventHandler(channel chan<- *model.BackendModuleEvent) resource.Handler {
	return func() resource.EventHandler  {
		return &BackendModuleEventHandler{channel: channel}
	}
}

type BackendModuleEventHandler struct {
	k8s     v1alpha1.BackendModule
	channel chan<- *model.BackendModuleEvent
}

func (h *BackendModuleEventHandler) K8sResource() interface{} {
	return &h.k8s
}

func (h *BackendModuleEventHandler) ShouldNotify() bool {
	return true
}

func (h *BackendModuleEventHandler) Notify(eventType model.EventType) {
	h.channel <- &model.BackendModuleEvent{
		Type:     &eventType,
		Resource: BackendModuleConverter{}.ToGQL(&h.k8s),
	}
}
