package shared

type Asset struct {
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	Parameters map[string]interface{} `json:"parameters"`
	Type       string                 `json:"type"`
	Files      []File                 `json:"files"`
	Status     AssetStatus            `json:"status"`
}

type AssetStatus struct {
	Phase   AssetPhaseType `json:"phase"`
	Reason  string         `json:"reason"`
	Message string         `json:"message"`
}

type AssetPhaseType string

const (
	AssetPhaseTypeReady AssetPhaseType = "READY"
)
