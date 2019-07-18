package assetstore

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/assetstore/upload"
)

const (
	docTopicDisplayNameFormat = "Documentation topic for service class id=%s"
	docTopicDescriptionFormat = "Documentation topic for service class id=%s"
)

const (
	documentationFileName = "content.json"
	openApiSpecFileName   = "apiSpec.json"
	eventsSpecFileName    = "asyncApiSpec.json"
	odataXMLSpecFileName  = "odata.xml"
	odataJSONSpecFileName = "odata.json"
	docsTopicLabelKey     = "cms.kyma-project.io/view-context"
	docsTopicLabelValue   = "service-catalog"
)

type Service interface {
	Put(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError
	Remove(id string) apperrors.AppError
}

type service struct {
	docsTopicRepository DocsTopicRepository
	uploadClient        upload.Client
}

func NewService(repository DocsTopicRepository, uploadClient upload.Client, insecureAssetDownload bool, assetstoreRequestTimeout int) Service {
	return &service{
		docsTopicRepository: repository,
		uploadClient:        uploadClient,
	}
}

func (s service) Put(id string, apiType docstopic.ApiType, documentation []byte, apiSpec []byte, eventsSpec []byte) apperrors.AppError {
	if documentation == nil && apiSpec == nil && eventsSpec == nil {
		return nil
	}

	docsTopic, err := s.createDocumentationTopic(id, apiType, documentation, apiSpec, eventsSpec)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}

	return s.docsTopicRepository.Upsert(docsTopic)
}

func (s service) Remove(id string) apperrors.AppError {
	return s.docsTopicRepository.Delete(id)
}

func (s service) createDocumentationTopic(id string, apiType docstopic.ApiType, documentation []byte, apiSpec []byte, eventsSpec []byte) (docstopic.Entry, apperrors.AppError) {
	docsTopic := docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(docTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(docTopicDescriptionFormat, id),
		Urls:        make(map[string]string),
		Labels:      map[string]string{docsTopicLabelKey: docsTopicLabelValue},
	}

	apiSpecFileName, apiSpecKey := getApiSpecFileNameAndKey(apiSpec, apiType)
	err := s.processSpec(apiSpec, apiSpecFileName, apiSpecKey, &docsTopic)
	if err != nil {
		return docstopic.Entry{}, err
	}

	err = s.processSpec(eventsSpec, eventsSpecFileName, docstopic.KeyAsyncApiSpec, &docsTopic)
	if err != nil {
		return docstopic.Entry{}, err
	}

	err = s.processSpec(documentation, documentationFileName, docstopic.KeyDocumentationSpec, &docsTopic)
	if err != nil {
		return docstopic.Entry{}, err
	}

	return docsTopic, nil
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
	if content != nil {
		outputFile, err := s.uploadClient.Upload(filename, content)
		if err != nil {
			return apperrors.Internal("Failed to upload file %s, %s.", filename, err)
		}

		docsTopicEntry.Urls[fileKey] = outputFile.RemotePath
	}

	return nil
}
