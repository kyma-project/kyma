package assetstore

import (
	"crypto/tls"
	"fmt"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/upload"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/download"
	"net/http"
	"time"
)

const (
	DocTopicDisplayNameFormat = "Documentation topic for service class id=%s"
	DocTopicDescriptionFormat = "Documentation topic for service class id=%s"
	specRequestTimeout        = time.Duration(5 * time.Second)
)

const (
	documentationFileName = "content.json"
	openApiSpecFileName   = "apiSpec.json"
	eventsSpecFileName    = "asyncApiSpec.json"
	odataXMLSpecFileName  = "odata.xml"
	odataJSONSpecFileName = "odata.json"
	DocsTopicLabelKey     = "cms.kyma-project.io/viewContext"
	DocsTopicLabelValue   = "service-catalog"
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

func NewService(repository DocsTopicRepository, uploadClient upload.Client, insecureAssetDownload bool) Service {
	downloadClient := download.NewClient(&http.Client{
		Timeout:   specRequestTimeout,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureAssetDownload}},
	})

	return &service{
		docsTopicRepository: repository,
		uploadClient:        uploadClient,
		downloadClient:      downloadClient,
	}
}

type ContentEntry struct {
	FileName string
	FileKey  string
	Content  []byte
}

func (s service) Put(id string, apiType docstopic.ApiType, documentation []byte, apiSpec []byte, eventsSpec []byte) apperrors.AppError {

	docsTopic, err := s.createDocumentationTopic(id, apiType, documentation, apiSpec, eventsSpec)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications: %s", err.Error())
	}

	return s.docsTopicRepository.Upsert(docsTopic)
}

func (s service) Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError) {

	docsTopic, err := s.docsTopicRepository.Get(id)
	if err != nil {
		return nil, nil, nil, apperrors.Internal("Failed to read Docs Topic.")
	}

	if docsTopic.Status != docstopic.StatusReady {
		return nil, nil, nil, nil
	}

	apiSpec, err = s.getApiSpec(docsTopic)
	if err != nil {
		return nil, nil, nil, err
	}

	eventsSpec, err = s.getEventsSpec(docsTopic)
	if err != nil {
		return nil, nil, nil, err
	}

	documentation, err = s.getDocumentation(docsTopic)
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
		DisplayName: fmt.Sprintf(DocTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(DocTopicDescriptionFormat, id),
		Urls:        make(map[string]string),
		Labels:      map[string]string{DocsTopicLabelKey: DocsTopicLabelValue},
	}

	apiSpecFileName, apiSpecKey := getApiSpecFileNameAndKey(apiSpec, apiType)
	err := s.processSpec(apiSpec, apiSpecFileName, apiSpecKey, &docsTopic)
	if err != nil {
		return docstopic.Entry{}, err
	}

	err = s.processSpec(eventsSpec, eventsSpecFileName, docstopic.KeyEventsSpec, &docsTopic)
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
		mimeType := http.DetectContentType(content)
		if mimeType == httpconsts.ContentTypeXML {
			return odataXMLSpecFileName, docstopic.KeyODataXMLSpec
		} else {
			return odataJSONSpecFileName, docstopic.KeyODataJSONSpec
		}

	} else {
		return openApiSpecFileName, docstopic.KeyOpenApiSpec
	}
}

func (s service) processSpec(content []byte, filename, fileKey string, docsTopicEntry *docstopic.Entry) apperrors.AppError {
	if content != nil {
		outputFile, err := s.uploadClient.Upload(filename, content)
		if err != nil {
			return apperrors.Internal("Failed to upload file: %s.", filename)
		}

		docsTopicEntry.Urls[fileKey] = outputFile.RemotePath
	}

	return nil
}

func (s service) getApiSpec(entry docstopic.Entry) ([]byte, apperrors.AppError) {

	url, found := entry.Urls[docstopic.KeyOpenApiSpec]
	if found {
		return s.downloadClient.Fetch(url)
	}

	url, found = entry.Urls[docstopic.KeyODataJSONSpec]
	if found {
		return s.downloadClient.Fetch(url)
	}

	url, found = entry.Urls[docstopic.KeyODataXMLSpec]
	if found {
		return s.downloadClient.Fetch(url)
	}

	return nil, nil
}

func (s service) getEventsSpec(entry docstopic.Entry) ([]byte, apperrors.AppError) {

	url, found := entry.Urls[docstopic.KeyEventsSpec]
	if found {
		return s.downloadClient.Fetch(url)
	}

	return nil, nil
}

func (s service) getDocumentation(entry docstopic.Entry) ([]byte, apperrors.AppError) {

	url, found := entry.Urls[docstopic.KeyDocumentationSpec]
	if found {
		return s.downloadClient.Fetch(url)
	}

	return nil, nil
}
