package v1alpha2

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"strconv"
	"strings"
)

const (
	defaultMaxInFlightMessages = "10"
	minEventTypeSegments       = 2
	invalidPrefix              = "sap.kyma.custom"
)

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
	if r.Spec.TypeMatching == "" {
		r.Spec.TypeMatching = TypeMatchingStandard
	}
	if r.Spec.Config[MaxInFlightMessages] == "" {
		if r.Spec.Config == nil {
			r.Spec.Config = map[string]string{}
		}
		r.Spec.Config[MaxInFlightMessages] = defaultMaxInFlightMessages
	}
}

//+kubebuilder:webhook:path=/validate-eventing-kyma-project-io-v1alpha2-subscription,mutating=false,failurePolicy=fail,sideEffects=None,groups=eventing.kyma-project.io,resources=subscriptions,verbs=create;update,versions=v1alpha2,name=vsubscription.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Subscription{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateCreate() error {
	logger.Info("validate create", "name", r.Name)
	return r.validateSubscription()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateUpdate(_ runtime.Object) error {
	logger.Info("validate update", "name", r.Name)
	return r.validateSubscription()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateDelete() error {
	logger.Info("validate delete", "name", r.Name)
	return nil
}

func (r *Subscription) validateSubscription() error {
	var allErrs field.ErrorList

	if err := r.validateSubscriptionSource(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateSubscriptionTypes(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateSubscriptionConfig(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateSubscriptionSink(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "eventing.kyma-project.io", Kind: "Subscription"},
		r.Name, allErrs)
}

func (r *Subscription) validateSubscriptionSource() *field.Error {
	if r.Spec.Source == "" && r.Spec.TypeMatching != TypeMatchingExact {
		return field.Invalid(field.NewPath("spec").Child("source"),
			r.Name, "must not be empty")
	}
	return nil
}

func (r *Subscription) validateSubscriptionTypes() *field.Error {
	if r.Spec.Types == nil || len(r.Spec.Types) == 0 {
		return field.Invalid(field.NewPath("spec").Child("types"),
			r.Name, "must not be empty")
	}
	if len(r.GetUniqueTypes()) != len(r.Spec.Types) {
		return field.Invalid(field.NewPath("spec").Child("types", "type"),
			r.Name, "must not have duplicate types")
	}
	for _, etype := range r.Spec.Types {
		if len(etype) == 0 {
			return field.Invalid(field.NewPath("spec").Child("types"),
				r.Name, "must not be of length zero")
		}
		if segments := strings.Split(etype, "."); len(segments) < minEventTypeSegments {
			return field.Invalid(field.NewPath("spec").Child("types"),
				r.Name, "must have minimum "+strconv.Itoa(minEventTypeSegments)+" segments")
		}
		if r.Spec.TypeMatching != TypeMatchingExact && strings.HasPrefix(etype, invalidPrefix) {
			return field.Invalid(field.NewPath("spec").Child("types"),
				r.Name, "must not have "+invalidPrefix+" as type prefix")
		}
	}
	return nil
}

func (r *Subscription) validateSubscriptionConfig() *field.Error {
	if isNotInt(r.Spec.Config[MaxInFlightMessages]) {
		return field.Invalid(field.NewPath("spec").Child("config"),
			r.Name, MaxInFlightMessages+" must be a stringified int value")
	}
	return nil
}

func (r *Subscription) validateSubscriptionSink() *field.Error {
	//if err := validator.ValidateSemantics(r); err != nil {
	//	return field.Invalid(field.NewPath("spec").Child("sink"),
	//		r.Name, err.Error())
	//}
	return nil
}

func isNotInt(value string) bool {
	if _, err := strconv.Atoi(value); err != nil {
		return true
	}
	return false
}
