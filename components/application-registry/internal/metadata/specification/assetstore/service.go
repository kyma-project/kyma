package assetstore

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/upload"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/download"
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
	Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError)
	Remove(id string) apperrors.AppError
}

type service struct {
	docsTopicRepository DocsTopicRepository
	uploadClient        upload.Client
	downloadClient      download.Client
}

func NewService(repository DocsTopicRepository, uploadClient upload.Client, insecureAssetDownload bool, assetstoreRequestTimeout int) Service {
	downloadClient := download.NewClient(&http.Client{
		Timeout:   time.Duration(assetstoreRequestTimeout) * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureAssetDownload}},
	}, authorization.NewStrategyFactory(authorization.FactoryConfiguration{OAuthClientTimeout: assetstoreRequestTimeout}))

	return &service{
		docsTopicRepository: repository,
		uploadClient:        uploadClient,
		downloadClient:      downloadClient,
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

func (s service) Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError) {
	docsTopic, err := s.docsTopicRepository.Get(id)
	if err != nil && err.Code() != apperrors.CodeNotFound {
		return nil, nil, nil, apperrors.Internal("Failed to read Docs Topic, %s.", err)
	}

	if docsTopic.Status != docstopic.StatusReady {
		return nil, nil, nil, nil
	}

	apiSpec, err = s.getApiSpec(docsTopic)
	if err != nil {
		return nil, nil, nil, err
	}

	eventsSpec, err = s.getSpec(docsTopic, docstopic.KeyAsyncApiSpec)
	if err != nil {
		return nil, nil, nil, err
	}

	documentation, err = s.getSpec(docsTopic, docstopic.KeyDocumentationSpec)
	if err != nil {
		return nil, nil, nil, err
	}

	return documentation, apiSpec, eventsSpec, nil
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

func (s service) getApiSpec(entry docstopic.Entry) ([]byte, apperrors.AppError) {
	url, found := entry.Urls[docstopic.KeyOpenApiSpec]
	if found {
		return s.downloadClient.Fetch(url, nil, nil)
	}

	url, found = entry.Urls[docstopic.KeyODataSpec]
	if found {
		return s.downloadClient.Fetch(url, nil, nil)
	}

	return nil, nil
}

func (s service) getSpec(entry docstopic.Entry, key string) ([]byte, apperrors.AppError) {
	url, found := entry.Urls[key]
	if found {
		return s.downloadClient.Fetch(url, nil, nil)
	}

	return nil, nil
}
