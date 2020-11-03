package target

import (
	"fmt"

	"github.com/google/uuid"
	bindErr "github.com/kyma-project/kyma/components/binding/internal/error"
	"github.com/kyma-project/kyma/components/binding/internal/storage"
	"github.com/kyma-project/kyma/components/binding/internal/worker"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// TODO: change Handler methods when TargetKind logic will be ready

type Handler struct {
	client dynamic.Interface
	storage worker.TargetKindStorage
}

func NewHandler(client dynamic.Interface, storage worker.TargetKindStorage) *Handler {
	return &Handler{
		client:  client,
		storage: storage,
	}
}

func (h *Handler) AddLabel(b *v1alpha1.Binding) error {
	resourceData, err := h.storage.Get(storage.Kind(b.Spec.Target.Kind))
	if err != nil {
		return errors.Wrapf(err, "while getting Kind %s from storage", b.Spec.Target.Kind)
	}
	resource, err := h.getResource(b, resourceData)
	if err != nil {
		return errors.Wrapf(err, "while getting resource for Binding %s", b.Name)
	}
	labelsToApply := map[string]string{h.labelKey(b): uuid.New().String()}
	if err := h.ensureLabelsAreApplied(resource, labelsToApply, resourceData.LabelFields); err != nil {
		return errors.Wrap(err, "while ensuring labels are applied")
	}
	err = h.updateResource(resource, resourceData)
	if err != nil {
		return errors.Wrap(err, "while updating resource")
	}

	return nil
}

func (h *Handler) LabelExist(b *v1alpha1.Binding) (bool, error) {
	resourceData, err := h.storage.Get(storage.Kind(b.Spec.Target.Kind))
	if err != nil {
		return false, errors.Wrapf(err, "while getting Kind %s from storage", b.Spec.Target.Kind)
	}
	resource, err := h.getResource(b, resourceData)
	if err != nil {
		return false, errors.Wrapf(err, "while getting resource for Binding %s", b.Name)
	}

	resourceLabels, err := h.getResourceLabels(resource, resourceData.LabelFields)
	if err != nil {
		return false, errors.Wrapf(err, "while getting injected labels for Binding %s", b.Name)
	}
	for key, _ := range resourceLabels {
		if key == h.labelKey(b) {
			return true, nil
		}

	}
	return false, nil
}

func (h *Handler) RemoveOldAddNewLabel(b *v1alpha1.Binding) error {
	resourceData, err := h.storage.Get(storage.Kind(b.Spec.Target.Kind))
	if err != nil {
		return errors.Wrapf(err, "while getting Kind %s from storage", b.Spec.Target.Kind)
	}
	resource, err := h.getResource(b, resourceData)
	if err != nil {
		return errors.Wrapf(err, "while getting resource for Binding %s", b.Name)
	}

	resourceLabels, err := h.getResourceLabels(resource, resourceData.LabelFields)
	if err != nil {
		return errors.Wrapf(err, "while getting injected labels for Binding %s", b.Name)
	}
	for key, _ := range resourceLabels {
		if key == h.labelKey(b) {
			labelsToApply := map[string]string{h.labelKey(b): uuid.New().String()}
			if err := h.ensureLabelsAreApplied(resource, labelsToApply, resourceData.LabelFields); err != nil {
				return errors.Wrap(err, "while ensuring labels are applied")
			}
		}
	}
	labelsToApply := map[string]string{h.labelKey(b): uuid.New().String()}
	if err := h.ensureLabelsAreApplied(resource, labelsToApply, resourceData.LabelFields); err != nil {
		return errors.Wrap(err, "while ensuring labels are applied")
	}

	err = h.updateResource(resource, resourceData)
	if err != nil {
		return errors.Wrap(err, "while updating resource")
	}

	return nil
}

func (h *Handler) RemoveLabel(b *v1alpha1.Binding) error {
	resourceData, err := h.storage.Get(storage.Kind(b.Spec.Target.Kind))
	if err != nil {
		return errors.Wrapf(err, "while getting Kind %s from storage", b.Spec.Target.Kind)
	}
	resource, err := h.getResource(b, resourceData)
	if err != nil {
		return errors.Wrapf(err, "while getting resource for Binding %s", b.Name)
	}
	existingLabels, err := h.getResourceLabels(resource, resourceData.LabelFields)
	if err != nil {
		return errors.Wrapf(err, "while getting injected labels for Binding %s", b.Name)
	}
	labelsToDelete := make(map[string]string, 0)

	for key, value := range existingLabels {
		if key != h.labelKey(b) {
			continue
		}
		labelsToDelete[key] = value
	}
	if err := h.ensureLabelsAreDeleted(resource, labelsToDelete, resourceData.LabelFields); err != nil {
		return errors.Wrapf(err, "while trying to delete labels %v+")
	}

	err = h.updateResource(resource, resourceData)
	if err != nil {
		return errors.Wrap(err, "while updating resource")
	}

	return nil
}

func (h *Handler) getResource(b *v1alpha1.Binding, data *storage.ResourceData) (*unstructured.Unstructured, error) {
	resource, err := h.client.Resource(data.Schema).Namespace(b.Namespace).Get(b.Spec.Target.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "while getting resource")
	}
	return resource, nil
}

func (h *Handler) updateResource(resource *unstructured.Unstructured, data *storage.ResourceData) error {
	_, err := h.client.Resource(data.Schema).Namespace(resource.GetNamespace()).Update(resource, metav1.UpdateOptions{})
	if err != nil {
		return bindErr.AsTemporaryError(err, "while updating target resource %s %s in namespace %s", resource.GetKind(), resource.GetName(), resource.GetNamespace())
	}
	return nil
}

func (h *Handler) labelKey(b *v1alpha1.Binding) string {
	return fmt.Sprintf("%s-%s", v1alpha1.BindingLabelKey, b.Name)
}

func (h *Handler) getResourceLabels(res *unstructured.Unstructured, labelFields []string) (map[string]string, error) {
	labels, err := h.findOrCreateLabelsField(res, labelFields)
	if err != nil {
		return make(map[string]string, 0), err
	}

	result := make(map[string]string, 0)
	for key, value := range labels {
		strKey := fmt.Sprintf("%v", key)
		strValue := fmt.Sprintf("%v", value)

		result[strKey] = strValue
	}
	return result, nil
}

func (h *Handler) ensureLabelsAreApplied(res *unstructured.Unstructured, labelsToApply map[string]string, labelFields []string) error {
	labels, err := h.findOrCreateLabelsField(res, labelFields)
	if err != nil {
		return err
	}
	for k, v := range labelsToApply {
		labels[k] = v
	}
	return nil
}

func (h *Handler) ensureLabelsAreDeleted(res *unstructured.Unstructured, labelsToDelete map[string]string, labelFields []string) error {
	labels, err := h.findOrCreateLabelsField(res, labelFields)
	if err != nil {
		return err
	}

	for k := range labelsToDelete {
		delete(labels, k)
	}

	return nil
}

func (h *Handler) findOrCreateLabelsField(res *unstructured.Unstructured, labelFields []string) (map[string]interface{}, error) {
	var val interface{} = res.Object

	for i, field := range labelFields {
		if m, ok := val.(map[string]interface{}); i < len(labelFields) && ok {
			val, ok = m[field]
			if !ok {
				m[field] = map[string]interface{}{}
				val = m[field]
			}
		} else {
			return nil, fmt.Errorf("accessor error: %v is of the type %T, expected map[string]interface{}", val, val)
		}
	}

	result, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected type of field is map[string]string, but was %T", val)
	}
	return result, nil
}