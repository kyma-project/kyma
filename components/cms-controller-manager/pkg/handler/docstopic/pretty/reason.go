package pretty

type Reason string

const (
	AssetCreated         Reason = "AssetCreated"
	AssetCreationFailed  Reason = "AssetCreationFailed"
	AssetsCreationFailed Reason = "AssetsCreationFailed"
	AssetsListingFailed  Reason = "AssetsListingFailed"
	AssetDeleted         Reason = "AssetDeleted"
	AssetDeletionFailed  Reason = "AssetDeletionFailed"
	AssetsDeletionFailed Reason = "AssetsDeletionFailed"
	AssetUpdated         Reason = "AssetUpdated"

	AssetUpdateFailed  Reason = "AssetUpdateFailed"
	AssetsUpdateFailed Reason = "AssetsUpdateFailed"
	AssetsReady        Reason = "AssetsReady"
	WaitingForAssets   Reason = "WaitingForAssets"
	BucketError        Reason = "BucketError"
)

func (r Reason) String() string {
	return string(r)
}

func (r Reason) Message() string {
	switch r {
	case AssetCreated:
		return "Asset %s has been created"
	default:
		return ""
	}
}
