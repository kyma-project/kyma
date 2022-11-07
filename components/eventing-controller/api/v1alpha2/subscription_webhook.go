package v1alpha2

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var subscriptionlog = logf.Log.WithName("subscription-resource")

func (r *Subscription) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-eventing-kyma-project-io-v1alpha2-subscription,mutating=true,failurePolicy=fail,sideEffects=None,groups=eventing.kyma-project.io,resources=subscriptions,verbs=create;update,versions=v1alpha2,name=msubscription.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Subscription{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Subscription) Default() {
	subscriptionlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

//+kubebuilder:webhook:path=/validate-eventing-kyma-project-io-v1alpha2-subscription,mutating=false,failurePolicy=fail,sideEffects=None,groups=eventing.kyma-project.io,resources=subscriptions,verbs=create;update,versions=v1alpha2,name=vsubscription.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Subscription{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateCreate() error {
	subscriptionlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateUpdate(old runtime.Object) error {
	subscriptionlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateDelete() error {
	subscriptionlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
