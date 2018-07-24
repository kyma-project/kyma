package controller

import (
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/pretty"
	"github.com/pkg/errors"
	k8sSettings "k8s.io/api/settings/v1alpha1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientSettingsV1alpha1 "k8s.io/client-go/kubernetes/typed/settings/v1alpha1"
)

// PodPresetModifier provides functionality needed to create and delete PodPreset
type PodPresetModifier struct {
	settingsClient clientSettingsV1alpha1.SettingsV1alpha1Interface
}

// NewPodPresetModifier creates a new PodPresetModifier
func NewPodPresetModifier(settingsClient clientSettingsV1alpha1.SettingsV1alpha1Interface) *PodPresetModifier {
	return &PodPresetModifier{
		settingsClient: settingsClient,
	}
}

// UpsertPodPreset creates a new PodPreset or update it if needed
func (m *PodPresetModifier) UpsertPodPreset(podPreset *k8sSettings.PodPreset) error {
	// TODO consider to add support for `ownerReferences` and then use v1.IsControlledBy method
	_, err := m.settingsClient.PodPresets(podPreset.Namespace).Create(podPreset)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		if err := m.settingsClient.PodPresets(podPreset.Namespace).Delete(podPreset.Name, &metaV1.DeleteOptions{}); err != nil {
			return errors.Wrapf(err, "while deleting %s", pretty.PodPresetName(podPreset))
		}
		if _, err := m.settingsClient.PodPresets(podPreset.Namespace).Create(podPreset); err != nil {
			return errors.Wrapf(err, "while re-creating %s", pretty.PodPresetName(podPreset))
		}
	default:
		return errors.Wrapf(err, "while creating %s", pretty.PodPresetName(podPreset))
	}

	return nil
}

// EnsurePodPresetDeleted deletes a PodPreset if needed
func (m *PodPresetModifier) EnsurePodPresetDeleted(namespace, name string) error {
	err := m.settingsClient.PodPresets(namespace).Delete(name, &metaV1.DeleteOptions{})
	if err != nil && !apiErrors.IsNotFound(err) {
		return errors.Wrapf(err, "while deleting PodPreset %s in namespace %s", name, namespace)
	}
	return nil
}
