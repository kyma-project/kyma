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
	AssetUpdateFailed    Reason = "AssetUpdateFailed"
	AssetsUpdateFailed   Reason = "AssetsUpdateFailed"
	AssetsReady          Reason = "AssetsReady"
	WaitingForAssets     Reason = "WaitingForAssets"
	BucketError          Reason = "BucketError"
)

func (r Reason) String() string {
	return string(r)
}

func (r Reason) Message() string {
	switch r {
	case AssetCreated:
		return "Asset %s has been created"
	case AssetCreationFailed:
		return "Asset %s couldn't be created due to error %s"
	case AssetsCreationFailed:
		return "Assets couldn't be created due to error %s"
	case AssetsListingFailed:
		return "Assets couldn't be listed due to error %s"
	case AssetDeleted:
		return "Assets %s has been deleted"
	case AssetDeletionFailed:
		return "Assets %s couldn't be deleted due to error %s"
	case AssetsDeletionFailed:
		return "Assets couldn't be deleted due to error %s"
	case AssetUpdated:
		return "Asset %s has been updated"
	case AssetUpdateFailed:
		return "Asset %s couldn't be updated due to error %s"
	case AssetsUpdateFailed:
		return "Assets couldn't be updated due to error %s"
	case AssetsReady:
		return "Assets are ready to use"
	case WaitingForAssets:
		return "Waiting for assets to be in Ready phase"
	case BucketError:
		return "Couldn't ensure if bucket exist due to error %s"
	default:
		return ""
	}
}
