package assetstore

import (
	"fmt"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
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

type Service interface {
	Put(id string, documentation, apiSpec, eventsSpec *ContentEntry) apperrors.AppError
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

func (s service) Put(id string, documentation *ContentEntry, apiSpec *ContentEntry, eventsSpec *ContentEntry) apperrors.AppError {

	docsTopic, err := s.createDocumentationTopic(id, documentation, apiSpec, eventsSpec)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications")
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

func (s service) createDocumentationTopic(id string, documentation *ContentEntry, apiSpec *ContentEntry, eventsSpec *ContentEntry) (docstopic.Entry, apperrors.AppError) {

	docsTopic := docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(DocTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(DocTopicDescriptionFormat, id),
		Urls:        make(map[string]string),
	}

	err := s.processSpec(apiSpec, &docsTopic)
	if err != nil {
		return docstopic.Entry{}, err
	}

	err = s.processSpec(eventsSpec, &docsTopic)
	if err != nil {
		return docstopic.Entry{}, err
	}

	err = s.processSpec(documentation, &docsTopic)
	if err != nil {
		return docstopic.Entry{}, err
	}

	return docsTopic, nil
}

func (s service) processSpec(contentEntry *ContentEntry, docsTopicEntry *docstopic.Entry) apperrors.AppError {
	if contentEntry != nil {
		outputFile, err := s.uploadFile(docsTopicEntry.Id, contentEntry)
		if err != nil {
			return apperrors.Internal("Failed to upload file: %s.", contentEntry.FileName)
		}

		docsTopicEntry.Urls[contentEntry.FileKey] = outputFile.RemotePath
	}

	return nil
}

func (s service) uploadFile(id string, contentEntry *ContentEntry) (upload.OutputFile, apperrors.AppError) {
	inputFile := upload.InputFile{
		Directory: id,
		Name:      contentEntry.FileName,
		Contents:  contentEntry.Content,
	}
	return s.uploadClient.Upload(inputFile)
}

func (s service) getApiSpec(entry docstopic.Entry) ([]byte, apperrors.AppError) {

	url, found := entry.Urls[docstopic.DocsTopicKeyOpenApiSpec]
	if found {
		return s.fetchUrl(url)
	}

	url, found = entry.Urls[docstopic.DocsTopicKeyODataJSONSpec]
	if found {
		return s.fetchUrl(url)
	}

	url, found = entry.Urls[docstopic.DocsTopicKeyODataXMLSpec]
	if found {
		return s.fetchUrl(url)
	}

	return nil, nil
}

func (s service) getEventsSpec(entry docstopic.Entry) ([]byte, apperrors.AppError) {

	url, found := entry.Urls[docstopic.DocsTopicKeyEventsSpec]
	if found {
		return s.fetchUrl(url)
	}

	return nil, nil
}

func (s service) getDocumentation(entry docstopic.Entry) ([]byte, apperrors.AppError) {

	url, found := entry.Urls[docstopic.DocsTopicKeyDocumentationSpec]
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
