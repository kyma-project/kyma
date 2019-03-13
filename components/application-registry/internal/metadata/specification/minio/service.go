package minio

import (
	"fmt"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
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

	documentationFileFullName = "content.json"
	eventsSpecFileFullName    = "asyncApiSpec.json"

	apiSpecFileName = "apiSpec"

	jsonExtension  = ".json"
	odataExtension = ".xml"
)

var apiExtensions = [...]string{
	jsonExtension,
	odataExtension,
}

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

	apperr = s.create(id, documentationFileFullName, documentation)
	if apperr != nil {
		return apperr
	}

	apiSpecFileFullName := makeFileFullName(apiSpec, apiSpecFileName)
	apperr = s.create(id, apiSpecFileFullName, apiSpec)
	if apperr != nil {
		return apperr
	}

	apperr = s.create(id, eventsSpecFileFullName, eventsSpec)
	if apperr != nil {
		return apperr
	}

	return nil
}

func (s *service) Get(id string) ([]byte, []byte, []byte, apperrors.AppError) {
	documentation, apperr := s.repository.Get(bucketName, makeFilePath(id, documentationFileFullName))
	if apperr != nil {
		return nil, nil, nil, apperr
	}

	apiSpec, apperr := tryDifferentExtensions(s.getApiSpec, id)
	if apperr != nil {
		return nil, nil, nil, apperr
	}

	eventsSpec, apperr := s.repository.Get(bucketName, makeFilePath(id, eventsSpecFileFullName))
	if apperr != nil {
		return nil, nil, nil, apperr
	}

	return documentation, apiSpec, eventsSpec, nil
}

func (s *service) Remove(id string) apperrors.AppError {
	apperr := s.repository.Remove(bucketName, makeFilePath(id, documentationFileFullName))
	if apperr != nil {
		return apperr
	}

	_, apperr = tryDifferentExtensions(s.removeApiSpec, id)
	if apperr != nil {
		return apperr
	}

	apperr = s.repository.Remove(bucketName, makeFilePath(id, eventsSpecFileFullName))
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

func (s *service) getApiSpec(id, extension string) ([]byte, apperrors.AppError) {
	return s.repository.Get(bucketName, makeFilePath(id, fmt.Sprintf("%s%s", apiSpecFileName, extension)))
}

func (s *service) removeApiSpec(id, extension string) ([]byte, apperrors.AppError) {
	return nil, s.repository.Remove(bucketName, makeFilePath(id, fmt.Sprintf("%s%s", apiSpecFileName, extension)))
}

func tryDifferentExtensions(serviceOperation func(string, string) ([]byte, apperrors.AppError), id string) ([]byte, apperrors.AppError) {
	var apperr apperrors.AppError
	for _, extension := range apiExtensions {
		data, err := serviceOperation(id, extension)
		if err == nil {
			return data, nil
		}

		if apperr != nil {
			apperr.Append(err.Error(), apperr)
		} else {
			apperr = err
		}
	}
	return nil, apperr
}

func makeFilePath(id, fileName string) string {
	return fmt.Sprintf("%s/%s/%s", typeName, id, fileName)
}

func makeFileFullName(data []byte, fileName string) string {
	if len(data) > 0 && string([]rune(string(data))[0]) == "<" {
		return fmt.Sprintf("%s%s", fileName, odataExtension)
	}
	return fmt.Sprintf("%s%s", fileName, jsonExtension)
}
