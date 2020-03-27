// Copied from:
// https://github.com/kyma-incubator/api-gateway/blob/master/api/v1alpha1/apiRule_types.go

package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

//StatusCode .
type StatusCode string

const (
	//StatusOK .
	StatusOK StatusCode = "OK"
	//StatusSkipped .
	StatusSkipped StatusCode = "SKIPPED"
	//StatusError .
	StatusError StatusCode = "ERROR"
)

// APIRuleSpec defines the desired state of ApiRule
type APIRuleSpec struct {
	// Definition of the service to expose
	Service *Service `json:"service"`
	// Gateway to be used
	Gateway *string `json:"gateway"`
	//Rules represents collection of Rule to apply
	Rules []Rule `json:"rules"`
}

// APIRuleStatus defines the observed state of ApiRule
type APIRuleStatus struct {
	LastProcessedTime    *metav1.Time           `json:"lastProcessedTime,omitempty"`
	ObservedGeneration   int64                  `json:"observedGeneration,omitempty"`
	APIRuleStatus        *APIRuleResourceStatus `json:"APIRuleStatus,omitempty"`
	VirtualServiceStatus *APIRuleResourceStatus `json:"virtualServiceStatus,omitempty"`
	AccessRuleStatus     *APIRuleResourceStatus `json:"accessRuleStatus,omitempty"`
}

// APIRule is the Schema for the apis ApiRule
type APIRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIRuleSpec   `json:"spec,omitempty"`
	Status APIRuleStatus `json:"status,omitempty"`
}

// APIRuleList contains a list of ApiRule
type APIRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIRule `json:"items"`
}

//Service .
type Service struct {
	// Name of the service
	Name *string `json:"name"`
	// Port of the service to expose
	Port *uint32 `json:"port"`
	// URL on which the service will be visible
	Host *string `json:"host"`
	// Defines if the service is internal (in cluster) or external
	// +optional
	IsExternal *bool `json:"external,omitempty"`
}

//Rule .
type Rule struct {
	// Path to be exposed
	Path string `json:"path"`
	// Set of allowed HTTP methods
	Methods []string `json:"methods,omitempty"`
	// Set of access strategies for a single path
	AccessStrategies []*Authenticator `json:"accessStrategies"`
	// Mutators to be used
	// +optional
	Mutators []*Mutator `json:"mutators,omitempty"`
}

//APIRuleResourceStatus .
type APIRuleResourceStatus struct {
	Code        StatusCode `json:"code,omitempty"`
	Description string     `json:"desc,omitempty"`
}
