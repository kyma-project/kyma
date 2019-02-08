package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AssetSpec defines the desired state of Asset
type AssetSpec struct {
	Source    AssetSource    `json:"source"`
	BucketRef AssetBucketRef `json:"bucketRef,omitempty"`
}

type AssetBucketRef struct {
	Name string `json:"name"`
}

type AssetSource struct {
	// +kubebuilder:validation:Enum=single,package,index
	Mode AssetMode `json:"mode"`
	Url  string    `json:"url"`

	// +optional
	ValidationWebhookService []AssetWebhookService `json:"validationWebhookService,omitempty"`

	// +optional
	MutationWebhookService []AssetWebhookService `json:"mutationWebhookService,omitempty"`
}

type AssetMode string

const (
	AssetSingle  AssetMode = "single"
	AssetPackage AssetMode = "package"
	AssetIndex   AssetMode = "index"
)

type AssetWebhookService struct {
	Name      string                `json:"name,omitempty"`
	Namespace string                `json:"namespace,omitempty"`
	Endpoint  string                `json:"endpoint,omitempty"`
	Metadata  *runtime.RawExtension `json:"metadata,omitempty"`
}

// AssetStatus defines the observed state of Asset
type AssetStatus struct {
	Phase             AssetPhase     `json:"phase"`
	Message           string         `json:"message,omitempty"`
	Reason            string         `json:"reason,omitempty"`
	AssetRef          AssetStatusRef `json:"assetRef,omitempty"`
	LastHeartbeatTime metav1.Time    `json:"lastHeartbeatTime"`
}

type AssetPhase string

const (
	AssetReady   AssetPhase = "Ready"
	AssetPending AssetPhase = "Pending"
	AssetFailed  AssetPhase = "Failed"
)

type AssetStatusRef struct {
	BaseUrl string   `json:"baseUrl"`
	Assets  []string `json:"assets,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Asset is the Schema for the assets API
// +k8s:openapi-gen=true
type Asset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AssetSpec   `json:"spec,omitempty"`
	Status AssetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AssetList contains a list of Asset
type AssetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Asset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Asset{}, &AssetList{})
}
