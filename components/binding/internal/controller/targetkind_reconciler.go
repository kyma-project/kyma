package controller

import (
	"context"
	"time"

	bindErr "github.com/kyma-project/kyma/components/binding/internal/errors"
	bindingsv1alpha1 "github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TargetKindWorker interface {
	Process(*bindingsv1alpha1.TargetKind, log.FieldLogger) (*bindingsv1alpha1.TargetKind, error)
	RemoveProcess(*bindingsv1alpha1.TargetKind, log.FieldLogger) error
}

// TargetKindReconciler reconciles a TargetKind object
type TargetKindReconciler struct {
	client client.Client
	worker TargetKindWorker
	log    log.FieldLogger
	scheme *runtime.Scheme
}

func NewTargetKindReconciler(client client.Client, worker TargetKindWorker, log log.FieldLogger, scheme *runtime.Scheme) *TargetKindReconciler {
	return &TargetKindReconciler{
		client: client,
		worker: worker,
		log:    log,
		scheme: scheme,
	}
}

func (r *TargetKindReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	var targetKind bindingsv1alpha1.TargetKind
	err := r.client.Get(ctx, req.NamespacedName, &targetKind)
	if err != nil {
		r.log.Warnf("TargetKind %s not found during reconcile process: %s", req.NamespacedName, err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if targetKind.DeletionTimestamp != nil {
		r.log.Infof("start the removal TargetKind process: %s", req.NamespacedName)
		// TODO: handle removing corresponding Binding here
		err := r.worker.RemoveProcess(targetKind.DeepCopy(), r.log.WithField("TargetKind", req.NamespacedName))
		if err != nil {
			r.log.Errorf("cannot finish remove process for %s: %s", req.NamespacedName, err)
			return ctrl.Result{RequeueAfter: 1 * time.Second}, err
		}
		return ctrl.Result{}, nil
	}

	r.log.Infof("start the reconcile TargetKind process: %s", req.NamespacedName)
	updatedTargetKind, err := r.worker.Process(targetKind.DeepCopy(), r.log.WithField("TargetKind", req.NamespacedName))
	switch {
	case bindErr.IsTemporaryError(err):
		r.log.Errorf("TargetKind %s process failed with temporary error: %s", req.NamespacedName, err)
		return ctrl.Result{RequeueAfter: 3 * time.Second}, err
	case err != nil:
		r.log.Errorf("TargetKind %s process failed: %s", req.NamespacedName, err)
		return ctrl.Result{}, err
	}

	err = r.client.Update(ctx, updatedTargetKind)
	if err != nil {
		r.log.Errorf("cannot update TargetKind resource %s: %s", req.NamespacedName, err)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	return ctrl.Result{}, nil
}

func (r *TargetKindReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bindingsv1alpha1.TargetKind{}).
		Complete(r)
}
