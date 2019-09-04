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
	Reason             AssetReason    `json:"reason,omitempty"`
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

// +kubebuilder:validation:Enum=single;package;index
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

type AssetReason string

const (
	AssetPulled                         AssetReason = "Pulled"
	AssetPullingFailed                  AssetReason = "PullingFailed"
	AssetUploaded                       AssetReason = "Uploaded"
	AssetUploadFailed                   AssetReason = "UploadFailed"
	AssetBucketNotReady                 AssetReason = "BucketNotReady"
	AssetBucketError                    AssetReason = "BucketError"
	AssetMutated                        AssetReason = "Mutated"
	AssetMutationFailed                 AssetReason = "MutationFailed"
	AssetMutationError                  AssetReason = "MutationError"
	AssetMetadataExtracted              AssetReason = "MetadataExtracted"
	AssetMetadataExtractionFailed       AssetReason = "MetadataExtractionFailed"
	AssetValidated                      AssetReason = "Validated"
	AssetValidationFailed               AssetReason = "ValidationFailed"
	AssetValidationError                AssetReason = "ValidationError"
	AssetMissingContent                 AssetReason = "MissingContent"
	AssetRemoteContentVerificationError AssetReason = "RemoteContentVerificationError"
	AssetCleanupError                   AssetReason = "CleanupError"
	AssetCleaned                        AssetReason = "Cleaned"
	AssetScheduled                      AssetReason = "Scheduled"
)

func (r AssetReason) String() string {
	return string(r)
}

func (r AssetReason) Message() string {
	switch r {
	case AssetPulled:
		return "Asset content has been pulled"
	case AssetPullingFailed:
		return "Asset content pulling failed due to error %s"
	case AssetUploaded:
		return "Asset content has been uploaded"
	case AssetUploadFailed:
		return "Asset content uploading failed due to error %s"
	case AssetBucketNotReady:
		return "Referenced bucket is not ready"
	case AssetBucketError:
		return "Reading bucket status failed due to error %s"
	case AssetMutated:
		return "Asset content has been mutated"
	case AssetMutationFailed:
		return "Asset mutation failed due to %+v"
	case AssetMutationError:
		return "Asset mutation failed due to error %s"
	case AssetMetadataExtracted:
		return "Metadata has been extracted from asset content"
	case AssetMetadataExtractionFailed:
		return "Metadata extraction failed due to error %s"
	case AssetValidated:
		return "Asset content has been validated"
	case AssetValidationFailed:
		return "Asset validation failed due to %+v"
	case AssetValidationError:
		return "Asset validation failed due to error %s"
	case AssetMissingContent:
		return "Asset content has been removed from remote storage"
	case AssetRemoteContentVerificationError:
		return "Asset content verification failed due to error %s"
	case AssetCleanupError:
		return "Removing old asset content failed due to error %s"
	case AssetCleaned:
		return "Old asset content hes been removed"
	case AssetScheduled:
		return "Asset scheduled for processing"
	default:
		return ""
	}
}
