package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

// ProtocolSettings defines the CE protocol setting specification implementation
type ProtocolSettings struct {
	// ContentMode defines content mode for eventing based on BEB
	// +optional
	ContentMode *string `json:"contentMode,omitempty"`

	// ExemptHandshake defines whether exempt handshake for eventing based on BEB
	// +optional
	ExemptHandshake *bool `json:"exemptHandshake,omitempty"`

	// Qos defines quality of service for eventing based on BEB
	// +optional
	Qos *string `json:"qos,omitempty"`

	// WebhookAuth defines the Webhook called by an active subscription in BEB
	// +optional
	WebhookAuth *WebhookAuth `json:"webhookAuth,omitempty"`
}

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

type EmsSubscriptionStatus struct {
	// SubscriptionStatus defines the status of the Subscription
	// +optional
	SubscriptionStatus string `json:"subscriptionStatus,omitempty"`

	// SubscriptionStatusReason defines the reason of the status
	// +optional
	SubscriptionStatusReason string `json:"subscriptionStatusReason,omitempty"`

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
