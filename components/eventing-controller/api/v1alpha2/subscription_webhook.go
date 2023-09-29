package v1alpha2

import (
	"strconv"
	"strings"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"

	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	DefaultMaxInFlightMessages = "10"
	minEventTypeSegments       = 2
	subdomainSegments          = 5
	InvalidPrefix              = "sap.kyma.custom"
	ClusterLocalURLSuffix      = "svc.cluster.local"
	ValidSource                = "source"
)

func (s *Subscription) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(s).
		Complete()
}

//nolint: lll
//+kubebuilder:webhook:path=/mutate-eventing-kyma-project-io-v1alpha2-subscription,mutating=true,failurePolicy=fail,sideEffects=None,groups=eventing.kyma-project.io,resources=subscriptions,verbs=create;update,versions=v1alpha2,name=msubscription.kb.io,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &Subscription{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (s *Subscription) Default() {
	if s.Spec.TypeMatching == "" {
		s.Spec.TypeMatching = TypeMatchingStandard
	}
	if s.Spec.Config[MaxInFlightMessages] == "" {
		if s.Spec.Config == nil {
			s.Spec.Config = map[string]string{}
		}
		s.Spec.Config[MaxInFlightMessages] = DefaultMaxInFlightMessages
	}
}

//nolint: lll
//+kubebuilder:webhook:path=/validate-eventing-kyma-project-io-v1alpha2-subscription,mutating=false,failurePolicy=fail,sideEffects=None,groups=eventing.kyma-project.io,resources=subscriptions,verbs=create;update,versions=v1alpha2,name=vsubscription.kb.io,admissionReviewVersions=v1beta1

var _ webhook.Validator = &Subscription{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (s *Subscription) ValidateCreate() (admission.Warnings, error) {
	return s.ValidateSubscription()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (s *Subscription) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	return s.ValidateSubscription()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (s *Subscription) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (s *Subscription) ValidateSubscription() (admission.Warnings, error) {
	var allErrs field.ErrorList

	if err := s.validateSubscriptionSource(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := s.validateSubscriptionTypes(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := s.validateSubscriptionConfig(); err != nil {
		allErrs = append(allErrs, err...)
	}
	if err := s.validateSubscriptionSink(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(GroupKind, s.Name, allErrs)
}

func (s *Subscription) validateSubscriptionSource() *field.Error {
	if s.Spec.Source == "" && s.Spec.TypeMatching != TypeMatchingExact {
		return MakeInvalidFieldError(SourcePath, s.Name, EmptyErrDetail)
	}
	// Check only if the source is valid for the cloud event, with a valid event type.
	if IsInvalidCE(s.Spec.Source, "") {
		return MakeInvalidFieldError(SourcePath, s.Name, InvalidURIErrDetail)
	}
	return nil
}

func (s *Subscription) validateSubscriptionTypes() *field.Error {
	if s.Spec.Types == nil || len(s.Spec.Types) == 0 {
		return MakeInvalidFieldError(TypesPath, s.Name, EmptyErrDetail)
	}
	if len(s.GetUniqueTypes()) != len(s.Spec.Types) {
		return MakeInvalidFieldError(TypesPath, s.Name, DuplicateTypesErrDetail)
	}
	for _, etype := range s.Spec.Types {
		if len(etype) == 0 {
			return MakeInvalidFieldError(TypesPath, s.Name, LengthErrDetail)
		}
		if segments := strings.Split(etype, "."); len(segments) < minEventTypeSegments {
			return MakeInvalidFieldError(TypesPath, s.Name, MinSegmentErrDetail)
		}
		if s.Spec.TypeMatching != TypeMatchingExact && strings.HasPrefix(etype, InvalidPrefix) {
			return MakeInvalidFieldError(TypesPath, s.Name, InvalidPrefixErrDetail)
		}
		// Check only is the event type is valid for the cloud event, with a valid source.
		if IsInvalidCE(ValidSource, etype) {
			return MakeInvalidFieldError(TypesPath, s.Name, InvalidURIErrDetail)
		}
	}
	return nil
}

func (s *Subscription) validateSubscriptionConfig() field.ErrorList {
	var allErrs field.ErrorList
	if isNotInt(s.Spec.Config[MaxInFlightMessages]) {
		allErrs = append(allErrs, MakeInvalidFieldError(ConfigPath, s.Name, StringIntErrDetail))
	}
	if s.ifKeyExistsInConfig(ProtocolSettingsQos) && types.IsInvalidQoS(s.Spec.Config[ProtocolSettingsQos]) {
		allErrs = append(allErrs, MakeInvalidFieldError(ConfigPath, s.Name, InvalidQosErrDetail))
	}
	if s.ifKeyExistsInConfig(WebhookAuthType) && types.IsInvalidAuthType(s.Spec.Config[WebhookAuthType]) {
		allErrs = append(allErrs, MakeInvalidFieldError(ConfigPath, s.Name, InvalidAuthTypeErrDetail))
	}
	if s.ifKeyExistsInConfig(WebhookAuthGrantType) && types.IsInvalidGrantType(s.Spec.Config[WebhookAuthGrantType]) {
		allErrs = append(allErrs, MakeInvalidFieldError(ConfigPath, s.Name, InvalidGrantTypeErrDetail))
	}
	return allErrs
}

func (s *Subscription) validateSubscriptionSink() *field.Error {
	if s.Spec.Sink == "" {
		return MakeInvalidFieldError(SinkPath, s.Name, EmptyErrDetail)
	}

	if !utils.IsValidScheme(s.Spec.Sink) {
		return MakeInvalidFieldError(SinkPath, s.Name, MissingSchemeErrDetail)
	}

	trimmedHost, subDomains, err := utils.GetSinkData(s.Spec.Sink)
	if err != nil {
		return MakeInvalidFieldError(SinkPath, s.Name, err.Error())
	}

	// Validate sink URL is a cluster local URL.
	if !strings.HasSuffix(trimmedHost, ClusterLocalURLSuffix) {
		return MakeInvalidFieldError(SinkPath, s.Name, SuffixMissingErrDetail)
	}

	// We expected a sink in the format "service.namespace.svc.cluster.local".
	if len(subDomains) != subdomainSegments {
		return MakeInvalidFieldError(SinkPath, s.Name, SubDomainsErrDetail+trimmedHost)
	}

	// Assumption: Subscription CR and Subscriber should be deployed in the same namespace.
	svcNs := subDomains[1]
	if s.Namespace != svcNs {
		return MakeInvalidFieldError(NSPath, s.Name, NSMismatchErrDetail+svcNs)
	}

	return nil
}

func (s *Subscription) ifKeyExistsInConfig(key string) bool {
	_, ok := s.Spec.Config[key]
	return ok
}

func isNotInt(value string) bool {
	if _, err := strconv.Atoi(value); err != nil {
		return true
	}
	return false
}

func IsInvalidCE(source, eventType string) bool {
	if source == "" {
		return false
	}
	newEvent := utils.GetCloudEvent(eventType)
	newEvent.SetSource(source)
	err := newEvent.Validate()
	return err != nil
}
