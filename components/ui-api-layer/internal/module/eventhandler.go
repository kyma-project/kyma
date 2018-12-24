package module

import (
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
)

type eventHandler struct {
	module PluggableModule
}
func newEventHandler(module PluggableModule) *eventHandler {
	return &eventHandler{module:module}
}

func (h *eventHandler) OnAdd(obj interface{}) {
	resource, ok := obj.(*v1alpha1.BackendModule)
	h.printIncorrectTypeErrorIfShould(ok, obj)

	if !h.isAddOrDeleteEventRelatedToTheModule(resource) {
		return
	}

	if h.module.IsEnabled() {
		return
	}

	glog.Infof("Enabling module %s...", h.module.Name())
	err := h.module.Enable()
	printModuleErrorIfShould(err, h.module, "enabling")
}

func (h *eventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldResource, ok := oldObj.(*v1alpha1.BackendModule)
	h.printIncorrectTypeErrorIfShould(ok, oldObj)

	newResource, ok := newObj.(*v1alpha1.BackendModule)
	h.printIncorrectTypeErrorIfShould(ok, newObj)


	if !h.isUpdateEventRelatedToTheModule(oldResource, newResource) {
		return
	}

	moduleName := h.module.Name()
	if oldResource.Name == moduleName {
		glog.Infof("Disabling module %s...", moduleName)
		err := h.module.Disable()
		printModuleErrorIfShould(err, h.module, "disabling")
	} else if newResource.Name == moduleName {
		glog.Infof("Enabling module %s...", moduleName)
		err := h.module.Enable()
		printModuleErrorIfShould(err, h.module, "enabling")
	}
}

func (h *eventHandler) OnDelete(obj interface{}) {
	resource, ok := obj.(*v1alpha1.BackendModule)
	h.printIncorrectTypeErrorIfShould(ok, obj)

	if !h.isAddOrDeleteEventRelatedToTheModule(resource) {
		return
	}

	if !h.module.IsEnabled() {
		return
	}

	glog.Infof("Disabling module %s...", h.module.Name())
	err := h.module.Disable()
	printModuleErrorIfShould(err, h.module, "disabling")
}

func (h *eventHandler) isAddOrDeleteEventRelatedToTheModule(resource *v1alpha1.BackendModule) bool {
	if resource == nil || resource.Name != h.module.Name() {
		return false
	}

	return true
}

func (h *eventHandler) isUpdateEventRelatedToTheModule(oldResource, newResource *v1alpha1.BackendModule) bool {
	if oldResource == nil || newResource == nil || oldResource.Name == newResource.Name {
		return false
	}

	moduleName := h.module.Name()
	if oldResource.Name != moduleName || newResource.Name != moduleName {
		return false
	}

	return true
}

func (h *eventHandler) printIncorrectTypeErrorIfShould(ok bool, obj interface{}) {
	if ok {
		return
	}

	glog.Errorf("Incorrect item type: %T, should be: *BackendModule", obj)
}
