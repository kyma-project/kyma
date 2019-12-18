package rafter

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

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

//go:generate mockery -name=Service
type Service interface {
	Put(id string, apiType clusterassetgroup.ApiType, spec []byte, specFormat clusterassetgroup.SpecFormat, specCategory clusterassetgroup.SpecCategory) apperrors.AppError
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

func (s service) Put(id string, apiType clusterassetgroup.ApiType, spec []byte, specFormat clusterassetgroup.SpecFormat, specCategory clusterassetgroup.SpecCategory) apperrors.AppError {
	if len(spec) == 0 {
		return nil
	}

	existingHash, err := s.getExistingAssetHash(id)
	if err != nil {
		return err
	}

	newHash := calculateHash(spec)

	if existingHash == emptyHash {
		return s.create(id, apiType, spec, specFormat, specCategory, newHash)
	}

	if newHash != existingHash {
		return s.update(id, apiType, spec, specFormat, specCategory, newHash)
	}
	return nil
}

func (s service) getExistingAssetHash(id string) (string, apperrors.AppError) {
	entry, err := s.clusterAssetGroupRepository.Get(id)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return "", nil
		}

		return "", err
	}

	return entry.SpecHash, nil
}

func (s service) Delete(id string) apperrors.AppError {
	return s.clusterAssetGroupRepository.Delete(id)
}

func (s service) create(id string, apiType clusterassetgroup.ApiType, spec []byte, specFormat clusterassetgroup.SpecFormat, specCategory clusterassetgroup.SpecCategory, hash string) apperrors.AppError {
	assetGroup, err := s.createClusterAssetGroup(id, apiType, spec, specFormat, specCategory)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}
	assetGroup.SpecHash = hash
	return s.clusterAssetGroupRepository.Create(assetGroup)
}

func (s service) update(id string, apiType clusterassetgroup.ApiType, spec []byte, specFormat clusterassetgroup.SpecFormat, specCategory clusterassetgroup.SpecCategory, hash string) apperrors.AppError {
	assetGroup, err := s.createClusterAssetGroup(id, apiType, spec, specFormat, specCategory)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}

	assetGroup.SpecHash = hash

	return s.clusterAssetGroupRepository.Update(assetGroup)
}

func (s service) createClusterAssetGroup(id string, apiType clusterassetgroup.ApiType, spec []byte, specFormat clusterassetgroup.SpecFormat, category clusterassetgroup.SpecCategory) (clusterassetgroup.Entry, apperrors.AppError) {
	assetGroup := clusterassetgroup.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(clusterAssetGroupDisplayNameFormat, id),
		Description: fmt.Sprintf(clusterAssetGroupDescriptionFormat, id),
		Urls:        make(map[string]string),
		Labels:      map[string]string{clusterAssetGroupLabelKey: clusterAssetGroupLabelValue},
	}

	specCategories := []clusterassetgroup.SpecCategory{
		clusterassetgroup.ApiSpec,
		clusterassetgroup.EventApiSpec,
	}

	if contains(specCategories, category) {
		apiSpecFileName, apiSpecKey := getApiSpecFileNameAndKey(spec, specFormat, apiType)
		err := s.processSpec(spec, apiSpecFileName, apiSpecKey, &assetGroup)
		if err != nil {
			return clusterassetgroup.Entry{}, err
		}
		return assetGroup, nil
	}

	return clusterassetgroup.Entry{}, apperrors.WrongInput("Unknown spec category.")
}

func contains(specCategories []clusterassetgroup.SpecCategory, category clusterassetgroup.SpecCategory) bool {
	for _, c := range specCategories {
		if category == c {
			return true
		}
	}
	return false
}

func getApiSpecFileNameAndKey(content []byte, specFormat clusterassetgroup.SpecFormat, apiType clusterassetgroup.ApiType) (fileName, key string) {
	switch apiType {
	case clusterassetgroup.OpenApiType:
		return specFileName(openApiSpecFileName, specFormat), clusterassetgroup.KeyOpenApiSpec
	case clusterassetgroup.ODataApiType:
		return specFileName(odataSpecFileName, specFormat), clusterassetgroup.KeyODataSpec
	case clusterassetgroup.AsyncApi:
		return specFileName(eventsSpecFileName, specFormat), clusterassetgroup.KeyAsyncApiSpec
	default:
		return "", ""
	}
}

func specFileName(fileName string, specFormat clusterassetgroup.SpecFormat) string {
	return fmt.Sprintf("%s.%s", fileName, specFormat)
}

func (s service) processSpec(content []byte, filename, fileKey string, assetGroupEntry *clusterassetgroup.Entry) apperrors.AppError {
	outputFile, err := s.uploadClient.Upload(filename, content)
	if err != nil {
		return apperrors.Internal("Failed to upload file %s, %s.", filename, err)
	}
	assetGroupEntry.Urls[fileKey] = outputFile.RemotePath

	return nil
}

func calculateHash(content []byte) string {
	if content == nil {
		return emptyHash
	}
	sum := md5.Sum(content)
	return hex.EncodeToString(sum[:])
}
