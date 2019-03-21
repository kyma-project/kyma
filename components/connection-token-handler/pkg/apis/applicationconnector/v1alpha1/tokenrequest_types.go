package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TokenRequestState represents state of tokenrequest
type TokenRequestState string

const (
	// TokenRequestStateOK is used when successfully fetched token
	TokenRequestStateOK TokenRequestState = "OK"
	// TokenRequestStateERR is used when an error occured during token fetch
	TokenRequestStateERR TokenRequestState = "Error"
)

// TokenRequestStatus defines the observed state of TokenRequest
type TokenRequestStatus struct {
	Token       string            `json:"token"`
	URL         string            `json:"url"`
	ExpireAfter metav1.Time       `json:"expireAfter"`
	Application string            `json:"application"`
	State       TokenRequestState `json:"state"`
}

type ClusterContext struct {
	Tenant string `json:"tenant,omitempty"`
	Group  string `json:"group,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TokenRequest is the Schema for the tokenrequests API
// +k8s:openapi-gen=true
type TokenRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Context ClusterContext `json:"context,omitempty"`

	Status TokenRequestStatus `json:"status,omitempty"`
}

// ShouldExpire method returs true if tokenrequest expired
func (tr *TokenRequest) ShouldExpire() bool {
	if tr.Status.ExpireAfter == *new(metav1.Time) {
		return false
	}

	currentTime := metav1.Now()
	return tr.Status.ExpireAfter.Before(&currentTime)
}

// ShouldFetch method retuns true if tokenrequest requires fethcing new token
func (tr *TokenRequest) ShouldFetch() bool {
	return tr.Status.Token == ""
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TokenRequestList contains a list of TokenRequest
type TokenRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TokenRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TokenRequest{}, &TokenRequestList{})
}
