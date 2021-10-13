//go:build !local
package backend

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
)

// sets this reconciler as owner of obj
func (r *Reconciler) setAsOwnerReference(ctx context.Context, obj metav1.Object) error {
	controllerNamespacedName := types.NamespacedName{
		Namespace: deployment.ControllerNamespace,
		Name:      deployment.ControllerName,
	}
	var deploymentController appsv1.Deployment
	if err := r.Cache.Get(ctx, controllerNamespacedName, &deploymentController); err != nil {
		r.namedLogger().Errorw("get controller NamespacedName failed", "error", err)
		return err
	}
	references := []metav1.OwnerReference{
		*metav1.NewControllerRef(&deploymentController, schema.GroupVersionKind{
			Group:   appsv1.SchemeGroupVersion.Group,
			Version: appsv1.SchemeGroupVersion.Version,
			Kind:    "Deployment",
		}),
	}
	obj.SetOwnerReferences(references)
	return nil
}
