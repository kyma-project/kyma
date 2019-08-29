package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	// +optional
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`
}

type DocsTopicPhase string

const (
	DocsTopicPending DocsTopicPhase = "Pending"
	DocsTopicReady   DocsTopicPhase = "Ready"
	DocsTopicFailed  DocsTopicPhase = "Failed"
)

type CommonDocsTopicStatus struct {
	// +kubebuilder:validation:Enum=Pending,Ready,Failed
	Phase             DocsTopicPhase  `json:"phase"`
	Reason            DocsTopicReason `json:"reason,omitempty"`
	Message           string          `json:"message,omitempty"`
	LastHeartbeatTime metav1.Time     `json:"lastHeartbeatTime"`
}

type DocsTopicReason string

const (
	DocsTopicAssetCreated               DocsTopicReason = "AssetCreated"
	DocsTopicAssetCreationFailed        DocsTopicReason = "AssetCreationFailed"
	DocsTopicAssetsCreationFailed       DocsTopicReason = "AssetsCreationFailed"
	DocsTopicAssetsListingFailed        DocsTopicReason = "AssetsListingFailed"
	DocsTopicAssetDeleted               DocsTopicReason = "AssetDeleted"
	DocsTopicAssetDeletionFailed        DocsTopicReason = "AssetDeletionFailed"
	DocsTopicAssetsDeletionFailed       DocsTopicReason = "AssetsDeletionFailed"
	DocsTopicAssetUpdated               DocsTopicReason = "AssetUpdated"
	DocsTopicAssetUpdateFailed          DocsTopicReason = "AssetUpdateFailed"
	DocsTopicAssetsUpdateFailed         DocsTopicReason = "AssetsUpdateFailed"
	DocsTopicAssetsReady                DocsTopicReason = "AssetsReady"
	DocsTopicWaitingForAssets           DocsTopicReason = "WaitingForAssets"
	DocsTopicBucketError                DocsTopicReason = "BucketError"
	DocsTopicAssetsWebhookGetFailed     DocsTopicReason = "AssetsWebhookGetFailed"
	DocsTopicAssetsSpecValidationFailed DocsTopicReason = "AssetsSpecValidationFailed"
)

func (r DocsTopicReason) String() string {
	return string(r)
}

func (r DocsTopicReason) Message() string {
	switch r {
	case DocsTopicAssetCreated:
		return "Asset %s has been created"
	case DocsTopicAssetCreationFailed:
		return "Asset %s couldn't be created due to error %s"
	case DocsTopicAssetsCreationFailed:
		return "Assets couldn't be created due to error %s"
	case DocsTopicAssetsListingFailed:
		return "Assets couldn't be listed due to error %s"
	case DocsTopicAssetDeleted:
		return "Assets %s has been deleted"
	case DocsTopicAssetDeletionFailed:
		return "Assets %s couldn't be deleted due to error %s"
	case DocsTopicAssetsDeletionFailed:
		return "Assets couldn't be deleted due to error %s"
	case DocsTopicAssetUpdated:
		return "Asset %s has been updated"
	case DocsTopicAssetUpdateFailed:
		return "Asset %s couldn't be updated due to error %s"
	case DocsTopicAssetsUpdateFailed:
		return "Assets couldn't be updated due to error %s"
	case DocsTopicAssetsReady:
		return "Assets are ready to use"
	case DocsTopicWaitingForAssets:
		return "Waiting for assets to be in Ready phase"
	case DocsTopicBucketError:
		return "Couldn't ensure if bucket exist due to error %s"
	case DocsTopicAssetsWebhookGetFailed:
		return "Unable to get webhook configuration %s"
	case DocsTopicAssetsSpecValidationFailed:
		return "Invalid asset specification, %s"
	default:
		return ""
	}
}
