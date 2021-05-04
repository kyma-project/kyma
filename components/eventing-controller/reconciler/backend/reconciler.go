package controllers

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// This probably deserves a better name...
	BebBackendSecretLabelKey = "eventing-beb-backend-secret"
)

type BackendReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=eventingbackends,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=eventingbackends/status,verbs=get;update;patch

func (r *BackendReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	var secret v1.Secret

	if err := r.Client.Get(ctx, req.NamespacedName, &secret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If the label is removed what to do? first check if the removed secret is mentioned in the backend CR?
	// Don't we need a finalizer for that?
	if !isSecretLabeledAsBebBackend(&secret) {
		return ctrl.Result{}, nil
	}

	r.Log.Info("Found a secret labeled for BEB backend!")
	// TODO: make sure there is only one such secret. Use an indexer to query with the label.
	// TODO: create/update the EventingBackend CR

	return ctrl.Result{}, nil
}

func isSecretLabeledAsBebBackend(s *v1.Secret) bool {
	for k, v := range s.Labels {
		if k == BebBackendSecretLabelKey && v == "true" {
			return true
		}
	}
	return false
}

func (r *BackendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&v1.Secret{}).Complete(r)
}
