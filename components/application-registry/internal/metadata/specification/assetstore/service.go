package assetstore

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/upload"
)

type Service interface {
	Put(id string, documentation ContentEntry, ContentEntry, eventsSpec ContentEntry) apperrors.AppError
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

func (s service) Put(id string, documentation ContentEntry, apiSpec ContentEntry, eventsSpec ContentEntry) apperrors.AppError {

	docsTopic, err := s.uploadSpecs(id, documentation, apiSpec, eventsSpec)
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

func (s service) uploadSpecs(id string, documentation ContentEntry, apiSpec ContentEntry, eventsSpec ContentEntry) (docstopic.DocumentationTopic, apperrors.AppError) {

	return docstopic.DocumentationTopic{}, nil
}

func (s service) Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError) {
	return nil, nil, nil, nil
}

func (s service) Remove(id string) apperrors.AppError {
	return nil
}
