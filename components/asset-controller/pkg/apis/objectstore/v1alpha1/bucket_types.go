package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Bucket is the Schema for the buckets API
// +k8s:openapi-gen=true
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec   `json:"spec,omitempty"`
	Status BucketStatus `json:"status,omitempty"`
}

// BucketSpec defines the desired state of Bucket
type BucketSpec struct {
	Region string `json:"region"`

	// +kubebuilder:validation:Enum=none,readonly,readwrite,writeonly
	Policy BucketPolicy `json:"policy"`
}

type BucketPolicy string

const (
	BucketPolicyNone      BucketPolicy = "none"
	BucketPolicyReadOnly               = "readonly"
	BucketPolicyReadWrite              = "readwrite"
	BucketPolicyWriteOnly              = "writeonly"
)

// BucketStatus defines the observed state of Bucket
type BucketStatus struct {
	Phase   BucketPhase `json:"phase,omitempty"`
	Message string      `json:"message,omitempty"`
	Reason  string      `json:"reason,omitempty"`
}

type BucketPhase string

const (
	BucketCreated BucketPhase = "Created"
	BucketDeleted BucketPhase = "Deleted"
	BucketFailed  BucketPhase = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BucketList contains a list of Bucket
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bucket{}, &BucketList{})
}
