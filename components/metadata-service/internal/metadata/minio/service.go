package minio

import (
	"fmt"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
)

type Service interface {
	Put(id string, documentation []byte, apiSpec []byte, eventsSpec []byte) apperrors.AppError
	Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError)
	Remove(id string) apperrors.AppError
}

type service struct {
	repository Repository
}

const (
	bucketName = "content"
	typeName   = "service-class"

	documentationFileName = "content.json"
	apiSpecFileName       = "apiSpec.json"
	eventsSpecFileName    = "asyncApiSpec.json"
)

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) Put(id string, documentation []byte, apiSpec []byte, eventsSpec []byte) apperrors.AppError {
	apperr := s.Remove(id)
	if apperr != nil {
		return apperr
	}

	apperr = s.create(id, documentationFileName, documentation)
	if apperr != nil {
		return apperr
	}

	apperr = s.create(id, apiSpecFileName, apiSpec)
	if apperr != nil {
		return apperr
	}

	apperr = s.create(id, eventsSpecFileName, eventsSpec)
	if apperr != nil {
		return apperr
	}

	return nil
}

func (s *service) Get(id string) ([]byte, []byte, []byte, apperrors.AppError) {
	documentation, apperr := s.repository.Get(bucketName, makeFilePath(id, documentationFileName))
	if apperr != nil {
		return nil, nil, nil, apperr
	}

	apiSpec, apperr := s.repository.Get(bucketName, makeFilePath(id, apiSpecFileName))
	if apperr != nil {
		return nil, nil, nil, apperr
	}

	eventsSpec, apperr := s.repository.Get(bucketName, makeFilePath(id, eventsSpecFileName))
	if apperr != nil {
		return nil, nil, nil, apperr
	}

	return documentation, apiSpec, eventsSpec, nil
}

func (s *service) Remove(id string) apperrors.AppError {
	apperr := s.repository.Remove(bucketName, makeFilePath(id, documentationFileName))
	if apperr != nil {
		return apperr
	}

	apperr = s.repository.Remove(bucketName, makeFilePath(id, apiSpecFileName))
	if apperr != nil {
		return apperr
	}

	apperr = s.repository.Remove(bucketName, makeFilePath(id, eventsSpecFileName))
	if apperr != nil {
		return apperr
	}

	return nil
}

func (s *service) create(id, fileName string, content []byte) apperrors.AppError {
	if content != nil {
		path := makeFilePath(id, fileName)
		return s.repository.Put(bucketName, path, content)
	}

	return nil
}

func makeFilePath(id, fileName string) string {
	return fmt.Sprintf("%s/%s/%s", typeName, id, fileName)
}
