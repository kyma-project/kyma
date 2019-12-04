package pretty

type Kind int

const (
	AssetGroup Kind = iota
	AssetGroupType
	AssetGroups
	AssetGroupsType

	ClusterAssetGroup
	ClusterAssetGroupType
	ClusterAssetGroups
	ClusterAssetGroupsType

	Asset
	AssetType
	Assets
	AssetsType

	ClusterAsset
	ClusterAssetType
	ClusterAssets
	ClusterAssetsType

	File
	FileType
	Files
	FilesType
)

func (k Kind) String() string {
	switch k {
	case AssetGroup:
		return "Asset Group"
	case AssetGroupType:
		return "AssetGroup"
	case AssetGroups:
		return "Asset Groups"
	case AssetGroupsType:
		return "[]AssetGroup"
	case ClusterAssetGroup:
		return "Cluster Asset Group"
	case ClusterAssetGroupType:
		return "ClusterAssetGroup"
	case ClusterAssetGroups:
		return "Cluster Asset Groups"
	case ClusterAssetGroupsType:
		return "[]ClusterAssetGroup"
	case Asset:
		return "Asset"
	case AssetType:
		return "Asset"
	case Assets:
		return "Assets"
	case AssetsType:
		return "[]Asset"
	case ClusterAsset:
		return "Cluster Asset"
	case ClusterAssetType:
		return "ClusterAsset"
	case ClusterAssets:
		return "Cluster Assets"
	case ClusterAssetsType:
		return "[]ClusterAsset"
	case File:
		return "File"
	case FileType:
		return "File"
	case Files:
		return "Files"
	case FilesType:
		return "[]File"
	default:
		return ""
	}
}
