package shared

type AssetGroup struct {
	Name        string           `json:"name"`
	Namespace   string           `json:"namespace"`
	GroupName   string           `json:"groupName"`
	DisplayName string           `json:"displayName"`
	Description string           `json:"description"`
	Assets      []Asset          `json:"assets"`
	Status      AssetGroupStatus `json:"status"`
}

type AssetGroupStatus struct {
	Phase   AssetGroupPhaseType `json:"phase"`
	Reason  string              `json:"reason"`
	Message string              `json:"message"`
}

type AssetGroupPhaseType string

const (
	AssetGroupPhaseTypeReady AssetGroupPhaseType = "READY"
)
