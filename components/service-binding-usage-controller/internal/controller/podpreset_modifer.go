package controller

import (
	"context"

	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/pretty"
	svcatSettings "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/settings/v1alpha1"
	settingsv1alpha1 "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/settings/v1alpha1"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodPresetModifier provides functionality needed to create and delete PodPreset
type PodPresetModifier struct {
	settingsClient settingsv1alpha1.SettingsV1alpha1Interface
}

// NewPodPresetModifier creates a new PodPresetModifier
func NewPodPresetModifier(settingsClient settingsv1alpha1.SettingsV1alpha1Interface) *PodPresetModifier {
	return &PodPresetModifier{
		settingsClient: settingsClient,
	}
}

// UpsertPodPreset creates a new PodPreset or update it if needed
func (m *PodPresetModifier) UpsertPodPreset(podPreset *svcatSettings.PodPreset) error {
	// TODO consider to add support for `ownerReferences` and then use v1.IsControlledBy method
	_, err := m.settingsClient.PodPresets(podPreset.Namespace).Create(context.TODO(), podPreset, metaV1.CreateOptions{})
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		if err := m.settingsClient.PodPresets(podPreset.Namespace).Delete(context.TODO(), podPreset.Name, metaV1.DeleteOptions{}); err != nil {
			return errors.Wrapf(err, "while deleting %s", pretty.PodPresetName(podPreset))
		}
		if _, err := m.settingsClient.PodPresets(podPreset.Namespace).Create(context.TODO(), podPreset, metaV1.CreateOptions{}); err != nil {
			return errors.Wrapf(err, "while re-creating %s", pretty.PodPresetName(podPreset))
		}
	default:
		return errors.Wrapf(err, "while creating %s", pretty.PodPresetName(podPreset))
	}

	return nil
}

// EnsurePodPresetDeleted deletes a PodPreset if needed
func (m *PodPresetModifier) EnsurePodPresetDeleted(namespace, name string) error {
	err := m.settingsClient.PodPresets(namespace).Delete(context.TODO(), name, metaV1.DeleteOptions{})
	if err != nil && !apiErrors.IsNotFound(err) {
		return errors.Wrapf(err, "while deleting PodPreset %s in namespace %s", name, namespace)
	}
	return nil
}
