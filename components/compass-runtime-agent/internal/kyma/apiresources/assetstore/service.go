package assetstore

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/assetstore/docstopic"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/assetstore/upload"
)

const (
	docTopicDisplayNameFormat = "Documentation topic for service class id=%s"
	docTopicDescriptionFormat = "Documentation topic for service class id=%s"
)

const (
	openApiSpecFileName = "apiSpec"
	eventsSpecFileName  = "asyncApiSpec"
	odataSpecFileName   = "odata"
	docsTopicLabelKey   = "cms.kyma-project.io/view-context"
	docsTopicLabelValue = "service-catalog"
	emptyHash           = ""
)

type Service interface {
	Put(id string, apiType docstopic.ApiType, spec []byte, specFormat docstopic.SpecFormat, specCategory docstopic.SpecCategory) apperrors.AppError
	Delete(id string) apperrors.AppError
}

type service struct {
	docsTopicRepository DocsTopicRepository
	uploadClient        upload.Client
}

func NewService(repository DocsTopicRepository, uploadClient upload.Client) Service {
	return &service{
		docsTopicRepository: repository,
		uploadClient:        uploadClient,
	}
}

func (s service) Put(id string, apiType docstopic.ApiType, spec []byte, specFormat docstopic.SpecFormat, specCategory docstopic.SpecCategory) apperrors.AppError {
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
	entry, err := s.docsTopicRepository.Get(id)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return "", nil
		}

		return "", err
	}

	return entry.SpecHash, nil
}

func (s service) Delete(id string) apperrors.AppError {
	return s.docsTopicRepository.Delete(id)
}

func (s service) create(id string, apiType docstopic.ApiType, spec []byte, specFormat docstopic.SpecFormat, specCategory docstopic.SpecCategory, hash string) apperrors.AppError {
	docsTopic, err := s.createDocumentationTopic(id, apiType, spec, specFormat, specCategory)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}
	docsTopic.SpecHash = hash
	return s.docsTopicRepository.Create(docsTopic)
}

func (s service) update(id string, apiType docstopic.ApiType, spec []byte, specFormat docstopic.SpecFormat, specCategory docstopic.SpecCategory, hash string) apperrors.AppError {
	docsTopic, err := s.createDocumentationTopic(id, apiType, spec, specFormat, specCategory)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}

	docsTopic.SpecHash = hash

	return s.docsTopicRepository.Update(docsTopic)
}

func (s service) createDocumentationTopic(id string, apiType docstopic.ApiType, spec []byte, specFormat docstopic.SpecFormat, category docstopic.SpecCategory) (docstopic.Entry, apperrors.AppError) {
	docsTopic := docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(docTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(docTopicDescriptionFormat, id),
		Urls:        make(map[string]string),
		Labels:      map[string]string{docsTopicLabelKey: docsTopicLabelValue},
	}

	specCategories := []docstopic.SpecCategory{
		docstopic.ApiSpec,
		docstopic.EventApiSpec,
	}

	if contains(specCategories, category) {
		apiSpecFileName, apiSpecKey := getApiSpecFileNameAndKey(spec, specFormat, apiType)
		err := s.processSpec(spec, apiSpecFileName, apiSpecKey, &docsTopic)
		if err != nil {
			return docstopic.Entry{}, err
		}
		return docsTopic, nil
	}

	return docstopic.Entry{}, apperrors.WrongInput("Unknown spec category.")
}

func contains(specCategories []docstopic.SpecCategory, category docstopic.SpecCategory) bool {
	for _, c := range specCategories {
		if category == c {
			return true
		}
	}
	return false
}

func getApiSpecFileNameAndKey(content []byte, specFormat docstopic.SpecFormat, apiType docstopic.ApiType) (fileName, key string) {
	switch apiType {
	case docstopic.OpenApiType:
		return specFileName(openApiSpecFileName, specFormat), docstopic.KeyOpenApiSpec
	case docstopic.ODataApiType:
		return specFileName(odataSpecFileName, specFormat), docstopic.KeyODataSpec
	case docstopic.AsyncApi:
		return specFileName(eventsSpecFileName, specFormat), docstopic.KeyAsyncApiSpec
	default:
		return "", ""
	}
}

func specFileName(fileName string, specFormat docstopic.SpecFormat) string {
	return fmt.Sprintf("%s.%s", fileName, specFormat)
}

func (s service) processSpec(content []byte, filename, fileKey string, docsTopicEntry *docstopic.Entry) apperrors.AppError {
	outputFile, err := s.uploadClient.Upload(filename, content)
	if err != nil {
		return apperrors.Internal("Failed to upload file %s, %s.", filename, err)
	}
	docsTopicEntry.Urls[fileKey] = outputFile.RemotePath

	return nil
}

func calculateHash(content []byte) string {
	if content == nil {
		return emptyHash
	}
	sum := md5.Sum(content)
	return hex.EncodeToString(sum[:])
}
