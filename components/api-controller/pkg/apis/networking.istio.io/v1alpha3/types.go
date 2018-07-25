package v1alpha3

import (
	"fmt"

	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualService describes Istio VirtualService
type VirtualService struct {
	k8sMeta.TypeMeta   `json:",inline"`
	k8sMeta.ObjectMeta `json:"metadata,omitempty"`
	Spec               *VirtualServiceSpec `json:"spec"`
}

func (v *VirtualService) String() string {
	return fmt.Sprintf("{Namespace: %v, Name: %v, UID: %v, Spec: %v}", v.Namespace, v.Name, v.UID, v.Spec)
}

// VirtualServiceSpec is the spec for VirtualService resource
type VirtualServiceSpec struct {
	Hosts    []string     `json:"hosts"`
	Gateways []string     `json:"gateways"`
	Http     []*HTTPRoute `json:"http"`
}

func (s *VirtualServiceSpec) String() string {
	return fmt.Sprintf("{Hosts: %v, Gateways: %v, Http: %v}", s.Hosts, s.Gateways, s.Http)
}

type HTTPRoute struct {
	Match []*HTTPMatchRequest  `json:"match"`
	Route []*DestinationWeight `json:"route"`
}

func (o *HTTPRoute) String() string {
	return fmt.Sprintf("{Match: %v, Route: %v}", o.Match, o.Route)
}

type DestinationWeight struct {
	Destination *Destination `json:"destination"`
}

type Destination struct {
	Host string        `json:"host"`
	Port *PortSelector `json:"port,omitempty"`
}

type PortSelector struct {
	Number uint32 `json:"number"`
}

type HTTPMatchRequest struct {
	Uri *StringMatch `json:"uri"`
}

type StringMatch struct {
	Regex string `json:"regex"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualServiceList is a list of Rule resources
type VirtualServiceList struct {
	k8sMeta.TypeMeta `json:",inline"`
	k8sMeta.ListMeta `json:"metadata,omitempty"`
	Items            []VirtualService `json:"items"`
}
