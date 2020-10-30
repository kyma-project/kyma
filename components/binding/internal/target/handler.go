package target

import (
	"fmt"

	"github.com/google/uuid"
	bindErr "github.com/kyma-project/kyma/components/binding/internal/error"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// TODO: change Handler methods when TargetKind logic will be ready

type Handler struct {
	client dynamic.NamespaceableResourceInterface
	fields []string
}

func NewHandler(client dynamic.NamespaceableResourceInterface, fields []string) *Handler {
	return &Handler{client: client,
		fields: fields,
	}
}

func (h *Handler) AddLabel(b *v1alpha1.Binding) error {
	//ctx := context.Background()
	//
	//deployment, err := h.getDeployment(ctx, b) // rozne w zaleznosci od binding.Spec.Target
	//if err != nil {
	//	return err
	//}
	//
	//if deployment.Spec.Template.ObjectMeta.Labels == nil {
	//	deployment.Spec.Template.ObjectMeta.Labels = make(map[string]string, 0)
	//}
	//deployment.Spec.Template.ObjectMeta.Labels[h.labelKey(b)] = uuid.New().String()
	//
	//if err := h.updateDeployment(ctx, deployment); err != nil {
	//	return err
	//}

	return nil
}

func (h *Handler) AddLabelToResource(b *v1alpha1.Binding) error {
	resource, err := h.client.Namespace(b.Namespace).Get(b.Spec.Target.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "while getting resource")
	}

	labelsToApply := map[string]string{h.labelKey(b): uuid.New().String()}
	err = h.ensureLabelsAreApplied(resource, labelsToApply)
	if err != nil {
		return err
	}

	_, err = h.client.Namespace(resource.GetNamespace()).Update(resource, metav1.UpdateOptions{})
	if err != nil {
		return bindErr.AsTemporaryError(err, "while updating target resource %s %s in namespace %s", resource.GetKind(), resource.GetName(), resource.GetNamespace())
	}

	return nil
}

func (h *Handler) LabelExist(b *v1alpha1.Binding) (bool, error) {
	//ctx := context.Background()
	//
	//deployment, err := h.getDeployment(ctx, b)
	//if err != nil {
	//	return false, err
	//}
	//
	//for key, _ := range deployment.Spec.Template.ObjectMeta.Labels {
	//	if key == h.labelKey(b) {
	//		return true, nil
	//	}
	//}

	return false, nil
}

func (h *Handler) RemoveOldAddNewLabel(b *v1alpha1.Binding) error {
	//ctx := context.Background()
	//
	//deployment, err := h.getDeployment(ctx, b)
	//if err != nil {
	//	return err
	//}
	//
	//for key, _ := range deployment.Spec.Template.ObjectMeta.Labels {
	//	if key == h.labelKey(b) {
	//		deployment.Spec.Template.ObjectMeta.Labels[h.labelKey(b)] = uuid.New().String()
	//	}
	//}
	//if deployment.Spec.Template.ObjectMeta.Labels == nil {
	//	deployment.Spec.Template.ObjectMeta.Labels = make(map[string]string, 0)
	//}
	//deployment.Spec.Template.ObjectMeta.Labels[h.labelKey(b)] = uuid.New().String()
	//
	//if err := h.updateDeployment(ctx, deployment); err != nil {
	//	return err
	//}

	return nil
}

func (h *Handler) RemoveLabel(b *v1alpha1.Binding) error {
	//ctx := context.Background()
	//
	//deployment, err := h.getDeployment(ctx, b)
	//if err != nil {
	//	return err
	//}
	//
	//newLabels := make(map[string]string, 0)
	//for key, value := range deployment.Spec.Template.ObjectMeta.Labels {
	//	if key == h.labelKey(b) {
	//		continue
	//	}
	//	newLabels[key] = value
	//}
	//deployment.Spec.Template.ObjectMeta.Labels = newLabels
	//
	//if err := h.updateDeployment(ctx, deployment); err != nil {
	//	return err
	//}

	return nil
}

//func (h *Handler) getDeployment(ctx context.Context, b *v1alpha1.Binding) (*v1.Deployment, error) {
//	var deployment v1.Deployment
//
//	err := h.client.Get(ctx, client.ObjectKey{Name: b.Spec.Target.Name, Namespace: b.Namespace}, &deployment)
//	if err != nil {
//		return &deployment, errors.Wrap(err, "while getting target deployment by client")
//	}
//
//	return deployment.DeepCopy(), nil
//}
//
//func (h *Handler) updateDeployment(ctx context.Context, deployment *v1.Deployment) error {
//	err := h.client.Update(ctx, deployment)
//	if err != nil {
//		return bindErr.AsTemporaryError(err, "while updating target deployment by client")
//	}
//
//	return nil
//}

func (h *Handler) labelKey(b *v1alpha1.Binding) string {
	return fmt.Sprintf("%s-%s", v1alpha1.BindingLabelKey, b.Name)
}

func (h *Handler) ensureLabelsAreApplied(res *unstructured.Unstructured, labelsToApply map[string]string) error {
	labels, err := h.findOrCreateLabelsField(res)
	if err != nil {
		return err
	}
	for k, v := range labelsToApply {
		labels[k] = v
	}
	return nil
}

func (h *Handler) findOrCreateLabelsField(res *unstructured.Unstructured) (map[string]interface{}, error) {
	var val interface{} = res.Object

	for i, field := range h.fields {
		if m, ok := val.(map[string]interface{}); i < len(h.fields) && ok {
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