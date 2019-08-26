package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CommonAssetSpec defines the desired state of Asset
type CommonAssetSpec struct {
	Source    AssetSource    `json:"source"`
	BucketRef AssetBucketRef `json:"bucketRef,omitempty"`
	// +optional
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`
}

// CommonAssetStatus defines the observed state of Asset
type CommonAssetStatus struct {
	Phase              AssetPhase     `json:"phase"`
	Message            string         `json:"message,omitempty"`
	Reason             string         `json:"reason,omitempty"`
	AssetRef           AssetStatusRef `json:"assetRef,omitempty"`
	LastHeartbeatTime  metav1.Time    `json:"lastHeartbeatTime"`
	ObservedGeneration int64          `json:"observedGeneration"`
}

type AssetPhase string

const (
	AssetReady   AssetPhase = "Ready"
	AssetPending AssetPhase = "Pending"
	AssetFailed  AssetPhase = "Failed"
)

type AssetStatusRef struct {
	BaseURL string      `json:"baseUrl"`
	Files   []AssetFile `json:"files,omitempty"`
}

type AssetFile struct {
	Name     string                `json:"name"`
	Metadata *runtime.RawExtension `json:"metadata,omitempty"`
}

type WebhookService struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	// +optional
	Endpoint string `json:"endpoint,omitempty"`
	// +optional
	Filter string `json:"filter,omitempty"`
}

type AssetWebhookService struct {
	WebhookService `json:",inline"`
	Parameters     *runtime.RawExtension `json:"parameters,omitempty"`
}

type AssetMode string

const (
	AssetSingle  AssetMode = "single"
	AssetPackage AssetMode = "package"
	AssetIndex   AssetMode = "index"
)

type AssetBucketRef struct {
	Name string `json:"name"`
}

type AssetSource struct {
	// +kubebuilder:validation:Enum=single,package,index
	Mode AssetMode `json:"mode"`
	URL  string    `json:"url"`
	// +optional
	Filter string `json:"filter,omitempty"`

	// +optional
	ValidationWebhookService []AssetWebhookService `json:"validationWebhookService,omitempty"`

	// +optional
	MutationWebhookService []AssetWebhookService `json:"mutationWebhookService,omitempty"`

	// +optional
	MetadataWebhookService []WebhookService `json:"metadataWebhookService,omitempty"`
}
