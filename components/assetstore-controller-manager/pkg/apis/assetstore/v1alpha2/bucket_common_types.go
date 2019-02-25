package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CommonBucketSpec defines the desired state of Bucket
type CommonBucketSpec struct {
	// +kubebuilder:validation:Enum=,us-east-1,us-west-1,us-west-2,eu-west-1,eu-central-1,ap-southeast-1,ap-southeast-2,ap-northeast-1,sa-east-1
	// +optional
	Region BucketRegion `json:"region,omitempty"`

	// +kubebuilder:validation:Enum=,none,readonly,writeonly,readwrite
	// +optional
	Policy BucketPolicy `json:"policy,omitempty"`
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

type BucketPolicy string

const (
	BucketPolicyNone      BucketPolicy = "none"
	BucketPolicyReadOnly  BucketPolicy = "readonly"
	BucketPolicyWriteOnly BucketPolicy = "writeonly"
	BucketPolicyReadWrite BucketPolicy = "readwrite"
)

// CommonBucketStatus defines the observed state of Bucket
type CommonBucketStatus struct {
	Url                string      `json:"url,omitempty"`
	Phase              BucketPhase `json:"phase,omitempty"`
	Message            string      `json:"message,omitempty"`
	Reason             string      `json:"reason,omitempty"`
	RemoteName         string      `json:"remoteName,omitempty"`
	LastHeartbeatTime  metav1.Time `json:"lastHeartbeatTime,omitempty"`
	ObservedGeneration int64       `json:"observedGeneration"`
}

type BucketPhase string

const (
	// BucketReady means that the bucket has been successfully created
	BucketReady BucketPhase = "Ready"

	// BucketFailed means that the bucket couldn't be created or has been deleted manually
	BucketFailed BucketPhase = "Failed"
)
