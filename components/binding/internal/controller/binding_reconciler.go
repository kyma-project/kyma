package controller

import (
	"context"
	"time"

	bindErr "github.com/kyma-project/kyma/components/binding/internal/error"
	bindingsv1alpha1 "github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type BindingWorker interface {
	Process(*bindingsv1alpha1.Binding, log.FieldLogger) (*bindingsv1alpha1.Binding, error)
	RemoveProcess(*bindingsv1alpha1.Binding, log.FieldLogger) (*bindingsv1alpha1.Binding, error)
}

// BindingReconciler reconciles a Binding object
type BindingReconciler struct {
	client client.Client
	worker BindingWorker
	log    log.FieldLogger
	scheme *runtime.Scheme
}

func NewBindingReconciler(client client.Client, worker BindingWorker, log log.FieldLogger, scheme *runtime.Scheme) *BindingReconciler {
	return &BindingReconciler{
		client: client,
		worker: worker,
		log:    log,
		scheme: scheme,
	}
}

func (r *BindingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	var binding bindingsv1alpha1.Binding

	err := r.client.Get(ctx, req.NamespacedName, &binding)
	if err != nil {
		r.log.Warnf("Binding %s not found during reconcile process: %s", req.NamespacedName, err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if binding.DeletionTimestamp != nil {
		r.log.Infof("start the removal Binding process: %s", req.NamespacedName)
		updatedBinding, err := r.worker.RemoveProcess(binding.DeepCopy(), r.log.WithField("Binding", req.NamespacedName))
		if err != nil {
			r.log.Errorf("cannot finish remove process for %s: %s", req.NamespacedName, err)
			return ctrl.Result{RequeueAfter: 1 * time.Second}, err
		}

		err = r.client.Update(ctx, updatedBinding)
		if err != nil {
			r.log.Errorf("cannot update Binding resource %s: %s", req.NamespacedName, err)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}

		return ctrl.Result{}, nil
	}

	r.log.Infof("start the reconcile Binding process: %s", req.NamespacedName)
	updatedBinding, err := r.worker.Process(binding.DeepCopy(), r.log.WithField("Binding", req.NamespacedName))
	switch {
	case bindErr.IsTemporaryError(err):
		r.log.Errorf("Binding %s process failed with temporary error: %s", req.NamespacedName, err)
		return ctrl.Result{RequeueAfter: 3 * time.Second}, err
	case err != nil:
		r.log.Errorf("Binding %s process failed: %s", req.NamespacedName, err)
		return ctrl.Result{}, err
	}

	err = r.client.Update(ctx, updatedBinding)
	if err != nil {
		r.log.Errorf("cannot update Binding resource %s: %s", req.NamespacedName, err)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	return ctrl.Result{}, nil
}

func (r *BindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bindingsv1alpha1.Binding{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
