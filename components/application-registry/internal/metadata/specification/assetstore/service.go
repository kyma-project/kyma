package assetstore

import (
	"fmt"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/upload"
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
	docsTopicRepository docstopic.Repository
	uploadClient        upload.Client
}

func NewService(repository docstopic.Repository, uploadClient upload.Client) Service {
	return &service{
		docsTopicRepository: repository,
		uploadClient:        uploadClient,
	}
}

type ContentEntry struct {
	FileName string
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

	apiSpecEntry, err := s.uploadFileAndCreateSpecEntry(id, apiSpec, DocsTopicApiSpecKey)
	if err != nil {
		return docstopic.Entry{}, apperrors.Internal("Failed to upload specification file.")
	}

	eventsSpecEntry, err := s.uploadFileAndCreateSpecEntry(id, eventsSpec, DocsTopicEventsSpecKey)
	if err != nil {
		return docstopic.Entry{}, apperrors.Internal("Failed to upload events specification file.")
	}

	docsSpecEntry, err := s.uploadFileAndCreateSpecEntry(id, documentation, DocsTopicDocumentationSpecKey)
	if err != nil {
		return docstopic.Entry{}, apperrors.Internal("Failed to upload documentation file.")
	}

	return docstopic.Entry{
		Id:            id,
		DisplayName:   fmt.Sprintf(DocTopicDisplayNameFormat, id),
		Description:   fmt.Sprintf(DocTopicDescriptionFormat, id),
		ApiSpec:       apiSpecEntry,
		EventsSpec:    eventsSpecEntry,
		Documentation: docsSpecEntry,
	}, nil
}

func (s service) uploadFileAndCreateSpecEntry(id string, entry *ContentEntry, key string) (*docstopic.SpecEntry, apperrors.AppError) {
	if entry != nil {
		outputFile, err := s.uploadSpec(id, entry)
		if err != nil {
			return nil, err
		}

		return &docstopic.SpecEntry{
			Url: outputFile.RemotePath,
			Key: key,
		}, nil
	}
	return nil, nil
}

func (s service) uploadSpec(id string, entry *ContentEntry) (upload.OutputFile, apperrors.AppError) {
	inputFile := upload.InputFile{
		Directory: id,
		Name:      entry.FileName,
		Contents:  entry.Content,
	}
	return s.uploadClient.Upload(inputFile)
}

func (s service) Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError) {
	return nil, nil, nil, nil
}

func (s service) Remove(id string) apperrors.AppError {
	return nil
}
