package target

import (
	"context"
	"fmt"
	"strings"

	bindErr "github.com/kyma-project/kyma/components/binding/internal/error"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: change Handler methods when TargetKind logic will be ready

type Handler struct {
	client client.Client
	storage *KindStorage
}

func NewHandler(client client.Client, storage *KindStorage) *Handler {
	return &Handler{client: client, storage: storage}
}

func (h *Handler) AddLabel(b *v1alpha1.Binding) error {
	ctx := context.Background()

	deployment, err := h.getDeployment(ctx, b) // rozne w zaleznosci od binding.Spec.Target
	if err != nil {
		return err
	}

	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		deployment.Spec.Template.ObjectMeta.Labels = make(map[string]string, 0)
	}
	deployment.Spec.Template.ObjectMeta.Labels[h.labelKey(b)] = uuid.New().String()

	if err := h.updateDeployment(ctx, deployment); err != nil {
		return err
	}

	return nil
}

func (h *Handler) AddLabelToResource(b *v1alpha1.Binding) error {
	ctx := context.Background()

	targetKind, err := h.storage.Get(b.Spec.Target.Kind)
	if err != nil {
		return err
	}

	resource, err := h.getResource(ctx, b)
	if err != nil {
		return err
	}

	var val interface{} = resource.Object

	labelsPath := strings.Split(targetKind.Spec.LabelsPath, ".")


	for i, field := range labelsPath {
		if m, ok := val.(map[string]interface{}); i < len(labelsPath) && ok {
			val, ok = m[field]
			if !ok {
				m[field] = map[string]interface{}{}
				val = m[field]
			}
		} else {
			return fmt.Errorf("accessor error: %v is of the type %T, expected map[string]interface{}", val, val)
		}
	}

	_, ok := val.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected type of field is map[string]string, but was %T", val)
	}

	err = h.client.Update(ctx, resource)
	if err != nil {
		return bindErr.AsTemporaryError(err, "while updating target deployment by client")
	}

	return nil
}

func (h *Handler) LabelExist(b *v1alpha1.Binding) (bool, error) {
	ctx := context.Background()

	deployment, err := h.getDeployment(ctx, b)
	if err != nil {
		return false, err
	}

	for key, _ := range deployment.Spec.Template.ObjectMeta.Labels {
		if key == h.labelKey(b) {
			return true, nil
		}
	}

	return false, nil
}

func (h *Handler) RemoveOldAddNewLabel(b *v1alpha1.Binding) error {
	ctx := context.Background()

	deployment, err := h.getDeployment(ctx, b)
	if err != nil {
		return err
	}

	for key, _ := range deployment.Spec.Template.ObjectMeta.Labels {
		if key == h.labelKey(b) {
			deployment.Spec.Template.ObjectMeta.Labels[h.labelKey(b)] = uuid.New().String()
		}
	}
	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		deployment.Spec.Template.ObjectMeta.Labels = make(map[string]string, 0)
	}
	deployment.Spec.Template.ObjectMeta.Labels[h.labelKey(b)] = uuid.New().String()

	if err := h.updateDeployment(ctx, deployment); err != nil {
		return err
	}

	return nil
}

func (h *Handler) RemoveLabel(b *v1alpha1.Binding) error {
	ctx := context.Background()

	deployment, err := h.getDeployment(ctx, b)
	if err != nil {
		return err
	}

	newLabels := make(map[string]string, 0)
	for key, value := range deployment.Spec.Template.ObjectMeta.Labels {
		if key == h.labelKey(b) {
			continue
		}
		newLabels[key] = value
	}
	deployment.Spec.Template.ObjectMeta.Labels = newLabels

	if err := h.updateDeployment(ctx, deployment); err != nil {
		return err
	}

	return nil
}

func (h *Handler) getResource(ctx context.Context, b *v1alpha1.Binding) (*unstructured.Unstructured, error) {
	var resource unstructured.Unstructured

	err := h.client.Get(ctx, client.ObjectKey{Name: b.Spec.Target.Name, Namespace: b.Namespace}, &resource)
	if err != nil {
		return nil, err
	}

	return resource.DeepCopy(), nil
}

func (h *Handler) getDeployment(ctx context.Context, b *v1alpha1.Binding) (*v1.Deployment, error) {
	var deployment v1.Deployment

	err := h.client.Get(ctx, client.ObjectKey{Name: b.Spec.Target.Name, Namespace: b.Namespace}, &deployment)
	if err != nil {
		return &deployment, errors.Wrap(err, "while getting target deployment by client")
	}

	return deployment.DeepCopy(), nil
}

func (h *Handler) updateDeployment(ctx context.Context, deployment *v1.Deployment) error {
	err := h.client.Update(ctx, deployment)
	if err != nil {
		return bindErr.AsTemporaryError(err, "while updating target deployment by client")
	}

	return nil
}

func (h *Handler) labelKey(b *v1alpha1.Binding) string {
	return fmt.Sprintf("%s-%s", v1alpha1.BindingLabelKey, b.Name)
}
