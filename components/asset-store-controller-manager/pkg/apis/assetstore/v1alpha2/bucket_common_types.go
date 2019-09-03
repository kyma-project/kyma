package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CommonBucketSpec defines the desired state of Bucket
type CommonBucketSpec struct {
	// +optional
	Region BucketRegion `json:"region,omitempty"`

	// +optional
	Policy BucketPolicy `json:"policy,omitempty"`
}

// +kubebuilder:validation:Enum=us-east-1;us-west-1;us-west-2;eu-west-1;eu-central-1;ap-southeast-1;ap-southeast-2;ap-northeast-1;sa-east-1;""
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

// +kubebuilder:validation:Enum=none;readonly;writeonly;readwrite;""
type BucketPolicy string

const (
	BucketPolicyNone      BucketPolicy = "none"
	BucketPolicyReadOnly  BucketPolicy = "readonly"
	BucketPolicyWriteOnly BucketPolicy = "writeonly"
	BucketPolicyReadWrite BucketPolicy = "readwrite"
)

// CommonBucketStatus defines the observed state of Bucket
type CommonBucketStatus struct {
	URL                string       `json:"url,omitempty"`
	Phase              BucketPhase  `json:"phase,omitempty"`
	Message            string       `json:"message,omitempty"`
	Reason             BucketReason `json:"reason,omitempty"`
	RemoteName         string       `json:"remoteName,omitempty"`
	LastHeartbeatTime  metav1.Time  `json:"lastHeartbeatTime,omitempty"`
	ObservedGeneration int64        `json:"observedGeneration"`
}

type BucketPhase string

const (
	// BucketReady means that the bucket has been successfully created
	BucketReady BucketPhase = "Ready"

	// BucketFailed means that the bucket couldn't be created or has been deleted manually
	BucketFailed BucketPhase = "Failed"
)

type BucketReason string

const (
	BucketNotFound                 BucketReason = "BucketNotFound"
	BucketCreationFailure          BucketReason = "BucketCreationFailure"
	BucketVerificationFailure      BucketReason = "BucketVerificationFailure"
	BucketCreated                  BucketReason = "BucketCreated"
	BucketPolicyUpdated            BucketReason = "BucketPolicyUpdated"
	BucketPolicyUpdateFailed       BucketReason = "BucketPolicyUpdateFailed"
	BucketPolicyVerificationFailed BucketReason = "BucketPolicyVerificationFailed"
	BucketPolicyHasBeenChanged     BucketReason = "BucketPolicyHasBeenChanged"
)

func (r BucketReason) String() string {
	return string(r)
}

func (r BucketReason) Message() string {
	switch r {
	case BucketCreated:
		return "Bucket has been created"
	case BucketNotFound:
		return "Bucket %s doesn't exist anymore"
	case BucketCreationFailure:
		return "Bucket couldn't be created due to error %s"
	case BucketVerificationFailure:
		return "Bucket couldn't be verified due to error %s"
	case BucketPolicyUpdated:
		return "Bucket policy has been updated"
	case BucketPolicyUpdateFailed:
		return "Bucket policy couldn't be set due to error %s"
	case BucketPolicyVerificationFailed:
		return "Bucket policy couldn't be verified due to error %s"
	case BucketPolicyHasBeenChanged:
		return "Remote bucket policy has been changed"
	default:
		return ""
	}
}
