package assetstore

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/upload"
)

const (
	docTopicDisplayNameFormat = "Documentation topic for service class id=%s"
	docTopicDescriptionFormat = "Documentation topic for service class id=%s"
)

const (
	openApiSpecFileName   = "apiSpec.json"
	eventsSpecFileName    = "asyncApiSpec.json"
	odataXMLSpecFileName  = "odata.xml"
	odataJSONSpecFileName = "odata.json"
	docsTopicLabelKey     = "cms.kyma-project.io/view-context"
	docsTopicLabelValue   = "service-catalog"
	emptyHash             = ""
)

type Service interface {
	Put(id string, apiType docstopic.ApiType, spec []byte, specCategory docstopic.SpecCategory) apperrors.AppError
	Remove(id string) apperrors.AppError
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

func (s service) Put(id string, apiType docstopic.ApiType, spec []byte, specCategory docstopic.SpecCategory) apperrors.AppError {
	if spec == nil {
		return nil
	}

	existingHash, err := s.getExistingAssetHash(id)
	if err != nil {
		return err
	}

	newHash := calculateHash(spec)

	if existingHash == emptyHash {
		return s.create(id, apiType, spec, specCategory, newHash)
	}

	if newHash != existingHash {
		return s.update(id, apiType, spec, specCategory, newHash)
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

func (s service) Remove(id string) apperrors.AppError {
	return s.docsTopicRepository.Delete(id)
}

func (s service) create(id string, apiType docstopic.ApiType, spec []byte, specCategory docstopic.SpecCategory, hash string) apperrors.AppError {
	docsTopic, err := s.createDocumentationTopic(id, apiType, spec, specCategory)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}
	docsTopic.SpecHash = hash
	return s.docsTopicRepository.Create(docsTopic)
}

func (s service) update(id string, apiType docstopic.ApiType, spec []byte, specCategory docstopic.SpecCategory, hash string) apperrors.AppError {
	docsTopic, err := s.createDocumentationTopic(id, apiType, spec, specCategory)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}

	docsTopic.SpecHash = hash

	return s.docsTopicRepository.Update(docsTopic)
}

func (s service) createDocumentationTopic(id string, apiType docstopic.ApiType, spec []byte, category docstopic.SpecCategory) (docstopic.Entry, apperrors.AppError) {
	docsTopic := docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(docTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(docTopicDescriptionFormat, id),
		Urls:        make(map[string]string),
		Labels:      map[string]string{docsTopicLabelKey: docsTopicLabelValue},
	}

	if category == docstopic.ApiSpec {
		apiSpecFileName, apiSpecKey := getApiSpecFileNameAndKey(spec, apiType)
		err := s.processSpec(spec, apiSpecFileName, apiSpecKey, &docsTopic)
		if err != nil {
			return docstopic.Entry{}, err
		}
		return docsTopic, nil
	}

	if category == docstopic.EventApiSpec {
		err := s.processSpec(spec, eventsSpecFileName, docstopic.KeyAsyncApiSpec, &docsTopic)
		if err != nil {
			return docstopic.Entry{}, err
		}
		return docsTopic, nil
	}

	return docstopic.Entry{}, apperrors.WrongInput("Unknown spec category.")
}

func getApiSpecFileNameAndKey(content []byte, apiType docstopic.ApiType) (fileName, key string) {
	if apiType == docstopic.ODataApiType {
		if isXML(content) {
			return odataXMLSpecFileName, docstopic.KeyODataSpec
		}

		return odataJSONSpecFileName, docstopic.KeyODataSpec
	}

	return openApiSpecFileName, docstopic.KeyOpenApiSpec
}

func isXML(content []byte) bool {
	const snippetLength = 512

	length := len(content)
	var snippet string

	if length < snippetLength {
		snippet = string(content)
	} else {
		snippet = string(content[:snippetLength])
	}

	openingIndex := strings.Index(snippet, "<")
	closingIndex := strings.Index(snippet, ">")

	return openingIndex != -1 && openingIndex < closingIndex
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
