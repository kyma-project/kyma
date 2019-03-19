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
}

func NewService(repository docstopic.Repository, uplocadClient upload.Client) Service {
	return &service{}
}

type ContentEntry struct {
	FileName string
	Content  []byte
}

func (s service) Put(id string, documentation ContentEntry, apiSpec ContentEntry, eventsSpec ContentEntry) apperrors.AppError {
	return nil
}

func (s service) Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError) {
	return nil, nil, nil, nil
}

func (s service) Remove(id string) apperrors.AppError {
	return nil
}
