package v1alpha2

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"strconv"
)

const defaultMaxInFlightMesages = "10"

var logger = logf.Log.WithName("subscription-resource")

func (r *Subscription) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-eventing-kyma-project-io-v1alpha2-subscription,mutating=true,failurePolicy=fail,sideEffects=None,groups=eventing.kyma-project.io,resources=subscriptions,verbs=create;update,versions=v1alpha2,name=msubscription.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Subscription{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Subscription) Default() {
	logger.Info("sub", "typematching", r.Spec.TypeMatching, "maxIn", r.Spec.Config[MaxInFlightMessages])
	if r.Spec.TypeMatching == "" {
		r.Spec.TypeMatching = TypeMatchingStandard
	}
	if r.Spec.Config == nil {
		r.Spec.Config = map[string]string{}
	}
	if r.Spec.Config[MaxInFlightMessages] == "" {
		r.Spec.Config[MaxInFlightMessages] = defaultMaxInFlightMesages
	}
}

//+kubebuilder:webhook:path=/validate-eventing-kyma-project-io-v1alpha2-subscription,mutating=false,failurePolicy=fail,sideEffects=None,groups=eventing.kyma-project.io,resources=subscriptions,verbs=create;update,versions=v1alpha2,name=vsubscription.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Subscription{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateCreate() error {
	logger.Info("validate create", "name", r.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateUpdate(old runtime.Object) error {
	logger.Info("validate update", "name", r.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateDelete() error {
	logger.Info("validate delete", "name", r.Name)
	return nil
}

func isNotInt(value string) bool {
	if _, err := strconv.Atoi(value); err != nil {
		return true
	}
	return false
}
