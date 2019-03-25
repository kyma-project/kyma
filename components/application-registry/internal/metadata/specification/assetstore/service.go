package assetstore

import (
	"fmt"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/upload"
	"io/ioutil"
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
)

type Service interface {
	Put(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError
	Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError)
	Remove(id string) apperrors.AppError
}

type service struct {
	docsTopicRepository DocsTopicRepository
	uploadClient        upload.Client
	httpClient          http.Client
}

func NewService(repository DocsTopicRepository, uploadClient upload.Client) Service {
	return &service{
		docsTopicRepository: repository,
		uploadClient:        uploadClient,
		httpClient: http.Client{
			Timeout: specRequestTimeout,
		},
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

	err = s.docsTopicRepository.Update(docsTopic)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return s.docsTopicRepository.Create(docsTopic)
		} else {
			return err
		}
	}

	return nil
}

func (s service) Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError) {

	docsTopic, err := s.docsTopicRepository.Get(id)
	if err != nil {
		return nil, nil, nil, apperrors.Internal("Failed to read Docs Topic.")
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

func (s service) createDocumentationTopic(id string, apiType docstopic.ApiType, documentation []byte, apiSpec []byte, eventsSpec []byte) (docstopic.Entry, apperrors.AppError) {

	docsTopic := docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(DocTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(DocTopicDescriptionFormat, id),
		Urls:        make(map[string]string),
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
		outputFile, err := s.uploadFile(docsTopicEntry.Id, filename, content)
		if err != nil {
			return apperrors.Internal("Failed to upload file: %s.", filename)
		}

		docsTopicEntry.Urls[fileKey] = outputFile.RemotePath
	}

	return nil
}

func (s service) uploadFile(id string, fileName string, content []byte) (upload.UploadedFile, apperrors.AppError) {
	inputFile := upload.InputFile{
		Directory: id,
		Name:      fileName,
		Contents:  content,
	}
	return s.uploadClient.Upload(inputFile)
}

func (s service) getApiSpec(entry docstopic.Entry) ([]byte, apperrors.AppError) {

	url, found := entry.Urls[docstopic.KeyOpenApiSpec]
	if found {
		return s.fetchUrl(url)
	}

	url, found = entry.Urls[docstopic.KeyODataJSONSpec]
	if found {
		return s.fetchUrl(url)
	}

	url, found = entry.Urls[docstopic.KeyODataXMLSpec]
	if found {
		return s.fetchUrl(url)
	}

	return nil, nil
}

func (s service) getEventsSpec(entry docstopic.Entry) ([]byte, apperrors.AppError) {

	url, found := entry.Urls[docstopic.KeyEventsSpec]
	if found {
		return s.fetchUrl(url)
	}

	return nil, nil
}

func (s service) getDocumentation(entry docstopic.Entry) ([]byte, apperrors.AppError) {

	url, found := entry.Urls[docstopic.KeyDocumentationSpec]
	if found {
		return s.fetchUrl(url)
	}

	return nil, nil
}

func (s service) fetchUrl(url string) ([]byte, apperrors.AppError) {
	res, err := s.requestAPISpec(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, apperrors.Internal("Failed to fetch from Asset Store.")
	}

	{
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, apperrors.Internal("Failed to read response body from Asset Store.")
		}

		return bytes, nil
	}
}

func (s service) requestAPISpec(specUrl string) (*http.Response, apperrors.AppError) {
	req, err := http.NewRequest(http.MethodGet, specUrl, nil)
	if err != nil {
		return nil, apperrors.Internal("Creating request for fetching API spec from %s failed, %s", specUrl, err.Error())
	}

	response, err := s.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("Fetching API spec from %s failed, %s", specUrl, err.Error())
	}

	if response.StatusCode != http.StatusOK {
		return nil, apperrors.UpstreamServerCallFailed("Fetching API spec from %s failed with status %s", specUrl, response.Status)
	}

	return response, nil
}

func (s service) Remove(id string) apperrors.AppError {
	return nil
}
