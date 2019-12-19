package shared

import "github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"

type ClusterAsset struct {
	Name       v1beta1.AssetGroupSourceName `json:"name"`
	Parameters map[string]interface{}       `json:"parameters"`
	Type       v1beta1.AssetGroupSourceType `json:"type"`
	Files      []File                       `json:"files"`
	Status     AssetStatus                  `json:"status"`
}
