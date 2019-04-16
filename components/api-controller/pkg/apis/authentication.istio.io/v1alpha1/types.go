package v1alpha1

import (
	"fmt"

	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Policy describes Istio policy
type Policy struct {
	k8sMeta.TypeMeta   `json:",inline"`
	k8sMeta.ObjectMeta `json:"metadata,omitempty"`
	Spec               *PolicySpec `json:"spec"`
}

func (p *Policy) String() string {
	return fmt.Sprintf("{Namespace: %v, Name: %v, UID: %v, Spec: %v}", p.Namespace, p.Name, p.UID, p.Spec)
}

// PolicySpec is the spec for Policy resource
type PolicySpec struct {
	Targets          Targets          `json:"targets"`
	PrincipalBinding PrincipalBinding `json:"principalBinding"`
	Origins          Origins          `json:"origins,omitempty"`
	Peers            Peers            `json:"peers,omitempty"`
}

type PrincipalBinding string

const (
	UseOrigin PrincipalBinding = "USE_ORIGIN"
)

func (p *PolicySpec) String() string {
	return fmt.Sprintf("{Targets: %v, PrincipalBinding: %v, Origins: %v}", p.Targets, p.PrincipalBinding, p.Origins)
}

type Targets []*Target

type Target struct {
	Name string `json:"name"`
}

func (t *Target) String() string {
	return fmt.Sprintf("{Name: %s}", t.Name)
}

type Origins []*Origin

type Origin struct {
	Jwt *Jwt `json:"jwt"`
}

func (o *Origin) String() string {
	return fmt.Sprintf("{Jwt: %v}", o.Jwt)
}

type Peers []*Peer

type Peer struct {
	MTLS struct{} `json:"mtls"`
}

type Jwt struct {
	Issuer  string `json:"issuer"`
	JwksUri string `json:"jwksUri"`
}

func (j *Jwt) String() string {
	return fmt.Sprintf("{Issuer: %s, JwksUri: %s}", j.Issuer, j.JwksUri)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyList is a list of Rule resources
type PolicyList struct {
	k8sMeta.TypeMeta `json:",inline"`
	k8sMeta.ListMeta `json:"metadata,omitempty"`
	Items            []Policy `json:"items"`
}
