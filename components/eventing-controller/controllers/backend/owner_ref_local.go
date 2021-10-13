//go:build local
package backend

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// sets this reconciler as owner of obj
func (r *Reconciler) setAsOwnerReference(ctx context.Context, obj metav1.Object) error {
	return nil
}
