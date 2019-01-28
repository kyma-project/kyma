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
	// +kubebuilder:validation:Enum=,us-east-1,us-west-1,us-west-2,eu-west-1,eu-central-1,ap-southeast-1,ap-southeast-2,ap-northeast-1,sa-east-1
	// +optional
	Region BucketRegion `json:"region,omitempty"`

	// +optional
	Policy string `json:"policy,omitempty"`
}

type BucketRegion string

const (
	BucketRegionUSEast1      BucketRegion = "us-east-1"
	BucketRegionUSWest1                   = "us-west-1"
	BucketRegionUSWest2                   = "us-west-2"
	BucketRegionEUEast1                   = "eu-west-1"
	BucketRegionEUCentral1                = "eu-central-1"
	BucketRegionAPSoutheast1              = "ap-southeast-1"
	BucketRegionAPSoutheast2              = "ap-southeast-2"
	BucketRegionAPNortheast1              = "ap-northeast-1"
	BucketRegionSAEast1                   = "sa-east-1"
)

// BucketStatus defines the observed state of Bucket
type BucketStatus struct {
	Phase             BucketPhase `json:"phase,omitempty"`
	Message           string      `json:"message,omitempty"`
	Reason            string      `json:"reason,omitempty"`
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
}

type BucketPhase string

const (
	// BucketReady means that the bucket has been successfully created
	BucketReady BucketPhase = "Ready"

	// BucketFailed means that the bucket couldn't be created or has been deleted manually
	BucketFailed BucketPhase = "Failed"
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
