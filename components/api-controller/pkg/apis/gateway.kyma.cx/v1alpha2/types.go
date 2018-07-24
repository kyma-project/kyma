package v1alpha2

import (
	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/meta/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Api struct {
	k8sMeta.TypeMeta   `json:",inline"`
	k8sMeta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiSpec   `json:"spec"`
	Status ApiStatus `json:"status"`
}

type ApiSpec struct {
	Service        Service        `json:"service"`
	Hostname       string         `json:"hostname"`
	Authentication Authentication `json:"authentication"`
}

type Service struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type Authentication []AuthenticationRule

type AuthenticationRule struct {
	Type AuthenticationType `json:"type"`
	Jwt  JwtAuthentication  `json:"jwt"`
}

type AuthenticationType string

const (
	JwtType AuthenticationType = "JWT"
)

type JwtAuthentication struct {
	JwksUri string `json:"jwksUri"`
	Issuer  string `json:"issuer"`
}

type ApiStatus struct {
	AuthenticationStatus kymaMeta.GatewayResourceStatus `json:"authenticationStatus,omitempty"`
	IngressStatus        kymaMeta.GatewayResourceStatus `json:"ingressStatus,omitempty"`
}

func (s *ApiStatus) IsEmpty() bool {
	return s.IngressStatus.IsEmpty() && s.AuthenticationStatus.IsEmpty()
}

func (s *ApiStatus) IsDone() bool {
	return s.IngressStatus.IsDone() && s.AuthenticationStatus.IsDone()
}

func (s *ApiStatus) IsInProgress() bool {
	return s.IngressStatus.IsInProgress() || s.AuthenticationStatus.IsInProgress()
}

func (s *ApiStatus) IsError() bool {
	return s.IngressStatus.IsError() || s.AuthenticationStatus.IsError()
}

func (s *ApiStatus) SetInProgress() {
	s.AuthenticationStatus = kymaMeta.GatewayResourceStatus{
		Code: kymaMeta.InProgress,
	}
	s.IngressStatus = kymaMeta.GatewayResourceStatus{
		Code: kymaMeta.InProgress,
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ApiList struct {
	k8sMeta.TypeMeta `json:",inline"`
	k8sMeta.ListMeta `json:"metadata,omitempty"`

	Items []Api `json:"items"`
}
