package controller

import (
	"context"

	bindingsv1alpha1 "github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TargetKindReconciler reconciles a TargetKind object
type TargetKindReconciler struct {
	client client.Client
	log    log.FieldLogger
	scheme *runtime.Scheme
}

func NewTargetKindReconciler(client client.Client, log log.FieldLogger, scheme *runtime.Scheme) *TargetKindReconciler {
	return &TargetKindReconciler{
		client: client,
		log:    log,
		scheme: scheme,
	}
}

func (r *TargetKindReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()

	// TODO: implement logic

	return ctrl.Result{}, nil
}

func (r *TargetKindReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bindingsv1alpha1.TargetKind{}).
		Complete(r)
}
