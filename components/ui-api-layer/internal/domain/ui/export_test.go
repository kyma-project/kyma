package ui

import (
	idppresetv1alpha1 "github.com/kyma-project/kyma/components/idppreset/pkg/client/clientset/versioned/typed/ui/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

func (r *idpPresetResolver) SetIDPPresetConverter(c gqlIDPPresetConverter) {
	r.idpPresetConverter = c
}

func NewIDPPresetService(client idppresetv1alpha1.UiV1alpha1Interface, informer cache.SharedIndexInformer) *idpPresetService {
	return newIDPPresetService(client, informer)
}

func NewIDPPresetResolver(service idpPresetSvc) *idpPresetResolver {
	return newIDPPresetResolver(service)
}
