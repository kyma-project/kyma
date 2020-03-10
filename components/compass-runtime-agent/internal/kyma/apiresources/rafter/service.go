package rafter

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/sirupsen/logrus"

	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/upload"
)

const (
	clusterAssetGroupDisplayNameFormat = "Documentation topic for service class id=%s"
	clusterAssetGroupDescriptionFormat = "Documentation topic for service class id=%s"
)

const (
	openApiSpecFileName         = "apiSpec"
	eventsSpecFileName          = "asyncApiSpec"
	odataSpecFileName           = "odata"
	clusterAssetGroupLabelKey   = "rafter.kyma-project.io/view-context"
	clusterAssetGroupLabelValue = "service-catalog"
	emptyHash                   = ""
)

// Data Mapping
// APIPackage = Cluster Asset Group ; Asset Group Name = Package.ID ;
// APIDefinition, EventDefinition modelled as source, source.Name = APIDefinition.ID or EventDefinition.ID
// Changes:
// Current: Source.Name = "<name>-service id"
//go:generate mockery -name=Service
type Service interface {
	Put(id string, assets []clusterassetgroup.Asset) apperrors.AppError
	Delete(id string) apperrors.AppError
}

type service struct {
	clusterAssetGroupRepository ClusterAssetGroupRepository
	uploadClient                upload.Client
}

func NewService(repository ClusterAssetGroupRepository, uploadClient upload.Client) Service {
	return &service{
		clusterAssetGroupRepository: repository,
		uploadClient:                uploadClient,
	}
}

func (s service) Put(id string, assets []clusterassetgroup.Asset) apperrors.AppError {
	logrus.Infof("Inserting %d files for id=%s", len(assets), id)
	if len(assets) == 0 {
		return nil
	}

	existingEntry, exists, err := s.getExistingEntry(id)
	if err != nil {
		return err
	}

	if exists {
		if compareAssetsHash(existingEntry.Assets, assets) {
			logrus.Infof("Skipped updating specs for id=%s", id)
			return nil
		}

		logrus.Infof("Updating specs for id=%s", id)
		return s.update(id, assets)
	}

	calculateAssetHash(assets)

	return s.create(id, assets)

}

func compareAssetsHash(currentAssets []clusterassetgroup.Asset, newAssets []clusterassetgroup.Asset) bool {
	if len(currentAssets) != len(newAssets) {
		return false
	}

	findAssetFunc := func(assetToFind clusterassetgroup.Asset, assets []clusterassetgroup.Asset) bool {
		for _, asset := range assets {
			if asset.Name == assetToFind.Name && asset.SpecHash == assetToFind.SpecHash {
				return true
			}
		}

		return false
	}

	for _, currentAsset := range currentAssets {
		if findAssetFunc(currentAsset, newAssets) {
			return true
		}
	}

	return false
}

func calculateAssetHash(assets []clusterassetgroup.Asset) {
	for _, asset := range assets {
		setHast(&asset, calculateHash(asset.Content))
	}
}

func setHast(asset *clusterassetgroup.Asset, hash string) {
	asset.SpecHash = hash
}

func (s service) getExistingEntry(id string) (clusterassetgroup.Entry, bool, apperrors.AppError) {
	entry, err := s.clusterAssetGroupRepository.Get(id)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return clusterassetgroup.Entry{}, false, nil
		}

		return clusterassetgroup.Entry{}, false, err
	}

	return entry, true, nil
}

func (s service) Delete(id string) apperrors.AppError {
	return s.clusterAssetGroupRepository.Delete(id)
}

func (s service) create(id string, assets []clusterassetgroup.Asset) apperrors.AppError {
	assetGroup, err := s.createClusterAssetGroup(id, assets)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}

	return s.clusterAssetGroupRepository.Create(assetGroup)
}

func (s service) update(id string, assets []clusterassetgroup.Asset) apperrors.AppError {
	assetGroup, err := s.createClusterAssetGroup(id, assets)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}

	return s.clusterAssetGroupRepository.Update(assetGroup)
}

func (s service) createClusterAssetGroup(id string, assets []clusterassetgroup.Asset) (clusterassetgroup.Entry, apperrors.AppError) {

	for i := 0; i < len(assets); i++ {
		asset := &assets[i]
		fileName := getApiSpecFileName(asset.Format, asset.Type)
		err := s.uploadFile(assets[i].Content, fileName, asset)
		if err != nil {
			return clusterassetgroup.Entry{}, err
		}
	}

	return clusterassetgroup.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(clusterAssetGroupDisplayNameFormat, id),
		Description: fmt.Sprintf(clusterAssetGroupDescriptionFormat, id),
		Labels:      map[string]string{clusterAssetGroupLabelKey: clusterAssetGroupLabelValue},
		Assets:      assets,
	}, nil
}

func getApiSpecFileName(specFormat clusterassetgroup.SpecFormat, apiType clusterassetgroup.ApiType) string {
	switch apiType {
	case clusterassetgroup.OpenApiType:
		return specFileName(openApiSpecFileName, specFormat)
	case clusterassetgroup.ODataApiType:
		return specFileName(odataSpecFileName, specFormat)
	case clusterassetgroup.AsyncApi:
		return specFileName(eventsSpecFileName, specFormat)
	default:
		return ""
	}
}

func specFileName(fileName string, specFormat clusterassetgroup.SpecFormat) string {
	return fmt.Sprintf("%s.%s", fileName, specFormat)
}

func (s service) uploadFile(content []byte, filename string, asset *clusterassetgroup.Asset) apperrors.AppError {
	outputFile, err := s.uploadClient.Upload(filename, content)
	if err != nil {
		return apperrors.Internal("Failed to upload file %s, %s.", filename, err)
	}
	asset.Url = outputFile.RemotePath

	return nil
}

func calculateHash(content []byte) string {
	if content == nil {
		return emptyHash
	}
	sum := md5.Sum(content)
	return hex.EncodeToString(sum[:])
}
