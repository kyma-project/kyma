package assetstore

import (
	"fmt"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/upload"
)

const (
	DocTopicDisplayNameFormat     = "Documentation topic for service class id=%s"
	DocTopicDescriptionFormat     = "Documentation topic for service class id=%s"
	DocsTopicApiSpecKey           = "api"
	DocsTopicEventsSpecKey        = "events"
	DocsTopicDocumentationSpecKey = "documentation"
)

type Service interface {
	Put(id string, documentation, apiSpec, eventsSpec *ContentEntry) apperrors.AppError
	Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError)
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

func (s service) Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError) {
	return nil, nil, nil, nil
}

func (s service) Remove(id string) apperrors.AppError {
	return nil
}
