package v1alpha2

import (
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConditionType string

const (
	ConditionSubscribed         ConditionType = "Subscribed"
	ConditionSubscriptionActive ConditionType = "Subscription active"
	ConditionAPIRuleStatus      ConditionType = "APIRule status"
	ConditionWebhookCallStatus  ConditionType = "Webhook call status"

	ConditionPublisherProxyReady ConditionType = "Publisher Proxy Ready"
	ConditionControllerReady     ConditionType = "Subscription Controller Ready"
)

type Condition struct {
	Type               ConditionType          `json:"type,omitempty"`
	Status             corev1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
	Reason             ConditionReason        `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
}

type ConditionReason string

const (
	ConditionReasonSubscriptionCreated        ConditionReason = "BEB Subscription created"
	ConditionReasonSubscriptionCreationFailed ConditionReason = "BEB Subscription creation failed"
	ConditionReasonSubscriptionActive         ConditionReason = "BEB Subscription active"
	ConditionReasonSubscriptionNotActive      ConditionReason = "BEB Subscription not active"
	ConditionReasonSubscriptionDeleted        ConditionReason = "BEB Subscription deleted"
	ConditionReasonAPIRuleStatusReady         ConditionReason = "APIRule status ready"
	ConditionReasonAPIRuleStatusNotReady      ConditionReason = "APIRule status not ready"
	ConditionReasonNATSSubscriptionActive     ConditionReason = "NATS Subscription active"
	ConditionReasonNATSSubscriptionNotActive  ConditionReason = "NATS Subscription not active"
	ConditionReasonWebhookCallStatus          ConditionReason = "BEB Subscription webhook call no errors status"

	ConditionReasonSubscriptionControllerReady    ConditionReason = "Subscription controller started"
	ConditionReasonSubscriptionControllerNotReady ConditionReason = "Subscription controller not ready"
	ConditionReasonPublisherDeploymentReady       ConditionReason = "Publisher proxy deployment ready"
	ConditionReasonPublisherDeploymentNotReady    ConditionReason = "Publisher proxy deployment not ready"
	ConditionReasonBackendCRSyncFailed            ConditionReason = "Backend CR sync failed"
	ConditionReasonPublisherProxySyncFailed       ConditionReason = "Publisher Proxy deployment sync failed"
	ConditionReasonControllerStartFailed          ConditionReason = "Starting the controller failed"
	ConditionReasonControllerStopFailed           ConditionReason = "Stopping the controller failed"
	ConditionReasonOauth2ClientSyncFailed         ConditionReason = "Failed to sync OAuth2 Client Credentials"
	ConditionReasonPublisherProxySecretError      ConditionReason = "Publisher proxy secret sync failed"
	ConditionDuplicateSecrets                     ConditionReason = "Multiple eventing backend labeled secrets exist"
)

// ConditionToAlpha1Version //todo
func ConditionToAlpha1Version(condition Condition) v1alpha1.Condition {
	return v1alpha1.Condition{
		Type:               v1alpha1.ConditionType(condition.Type),
		Status:             condition.Status,
		LastTransitionTime: condition.LastTransitionTime,
		Reason:             v1alpha1.ConditionReason(condition.Reason),
		Message:            condition.Message,
	}
}

type EventType struct {
	OriginalType string `json:"originalType"`
	CleanType    string `json:"cleanType"`
}

// Backend contains Backend-specific fields
type Backend struct {
	// BEB-specific fields

	// Ev2hash defines the hash for the Subscription custom resource
	// +optional
	Ev2hash int64 `json:"ev2hash,omitempty"`

	// Emshash defines the hash for the Subscription in BEB
	// +optional
	Emshash int64 `json:"emshash,omitempty"`

	// ExternalSink defines the webhook URL which is used by BEB to trigger subscribers
	// +optional
	ExternalSink string `json:"externalSink,omitempty"`

	// FailedActivation defines the reason if a Subscription had failed activation in BEB
	// +optional
	FailedActivation string `json:"failedActivation,omitempty"`

	// APIRuleName defines the name of the APIRule which is used by the Subscription
	// +optional
	APIRuleName string `json:"apiRuleName,omitempty"`

	// EmsSubscriptionStatus defines the status of Subscription in BEB
	// +optional
	EmsSubscriptionStatus *EmsSubscriptionStatus `json:"emsSubscriptionStatus,omitempty"`

	// NATS-specific fields

	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxInFlightMessages int `json:"maxInFlightMessages,omitempty"`

	// +optional
	Types []JetStreamTypes `json:"types,omitempty"`
}

type EmsSubscriptionStatus struct {
	// Status defines the status of the Subscription
	// +optional
	Status string `json:"status,omitempty"`

	// StatusReason defines the reason of the status
	// +optional
	StatusReason string `json:"statusReason,omitempty"`

	// LastSuccessfulDelivery defines the timestamp of the last successful delivery
	// +optional
	LastSuccessfulDelivery string `json:"lastSuccessfulDelivery,omitempty"`

	// LastFailedDelivery defines the timestamp of the last failed delivery
	// +optional
	LastFailedDelivery string `json:"lastFailedDelivery,omitempty"`

	// LastFailedDeliveryReason defines the reason of failed delivery
	// +optional
	LastFailedDeliveryReason string `json:"lastFailedDeliveryReason,omitempty"`
}

// WebhookAuth defines the Webhook called by an active subscription in BEB
type WebhookAuth struct {
	// Type defines type of authentication
	// +optional
	Type string `json:"type,omitempty"`

	// GrantType defines grant type for OAuth2
	GrantType string `json:"grantType"`

	// ClientID defines clientID for OAuth2
	ClientID string `json:"clientId"`

	// ClientSecret defines client secret for OAuth2
	ClientSecret string `json:"clientSecret"`

	// TokenURL defines token URL for OAuth2
	TokenURL string `json:"tokenUrl"`

	// Scope defines scope for OAuth2
	Scope []string `json:"scope,omitempty"`
}

type JetStreamTypes struct {
	OriginalType string `json:"originalType"`
	ConsumerName string `json:"consumerName"`
}
