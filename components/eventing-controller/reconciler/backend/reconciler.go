package controllers

import (
	"context"
	"github.com/go-logr/logr"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// This probably deserves a better name...
	BebBackendSecretLabelKey   = "eventing-beb-backend-secret"
	DefaultEventingBackendName = "eventing-backend"
	// TODO: where to get this namespace
	DefaultEventingBackendNamespace = "kyma-system"
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

	if err := r.Get(ctx, req.NamespacedName, &secret); err != nil {
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

	backend, err := r.getOrCreateBackendCR(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	backend.Status = eventingv1alpha1.EventingBackendStatus{
		Backend:         BebBackendSecretLabelKey,
		ControllerReady: boolPtr(false),
		EventingReady:   boolPtr(false),
		PublisherReady:  boolPtr(false),
	}
	if err := r.Status().Update(ctx, backend); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *BackendReconciler) getOrCreateBackendCR(ctx context.Context) (*eventingv1alpha1.EventingBackend, error) {
	var backend eventingv1alpha1.EventingBackend
	defaultEventingBackend := types.NamespacedName{
		Namespace: DefaultEventingBackendNamespace,
		Name:      DefaultEventingBackendName,
	}
	err := r.Get(ctx, defaultEventingBackend, &backend)
	if err == nil {
		return &backend, nil
	}
	if errors.IsNotFound(err) {
		// if the CR doesn't exit, create one.
		backend = eventingv1alpha1.EventingBackend{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DefaultEventingBackendName,
				Namespace: DefaultEventingBackendNamespace,
			},
			Spec: eventingv1alpha1.EventingBackendSpec{},
		}
		if err := r.Create(ctx, &backend); err != nil {
			r.Log.Error(err, "Cannot create an EventingBackend CR")
			return nil, err
		}
		return &backend, nil
	}
	return nil, err
}

func isSecretLabeledAsBebBackend(s *v1.Secret) bool {
	for k, v := range s.Labels {
		if k == BebBackendSecretLabelKey && v == "true" {
			return true
		}
	}
	return false
}

func boolPtr(b bool) *bool {
	return &b
}

func (r *BackendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&v1.Secret{}).Complete(r)
}
