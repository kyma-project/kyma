package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CommonDocsTopicSpec struct {
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	// +kubebuilder:validation:MinItems=1
	Sources []Source `json:"sources"`
}

type DocsTopicMode string

const (
	DocsTopicSingle  DocsTopicMode = "single"
	DocsTopicPackage DocsTopicMode = "package"
	DocsTopicIndex   DocsTopicMode = "index"
)

type Source struct {
	// +kubebuilder:validation:Pattern=^[a-z][a-zA-Z0-9-]*[a-zA-Z0-9]$
	Name string `json:"name"`
	// +kubebuilder:validation:Pattern=^[a-z][a-zA-Z0-9\._-]*[a-zA-Z0-9]$
	Type string `json:"type"`
	URL  string `json:"url"`
	// +kubebuilder:validation:Enum=single,package,index
	Mode   DocsTopicMode `json:"mode"`
	Filter string        `json:"filter,omitempty"`
}

type DocsTopicPhase string

const (
	DocsTopicPending DocsTopicPhase = "Pending"
	DocsTopicReady   DocsTopicPhase = "Ready"
	DocsTopicFailed  DocsTopicPhase = "Failed"
)

type CommonDocsTopicStatus struct {
	// +kubebuilder:validation:Enum=Pending,Ready,Failed
	Phase             DocsTopicPhase `json:"phase"`
	Reason            string         `json:"reason,omitempty"`
	Message           string         `json:"message,omitempty"`
	LastHeartbeatTime metav1.Time    `json:"lastHeartbeatTime"`
}
