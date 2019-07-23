package synchronization

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
)

//go:generate mockery -name=Service
type Service interface {
	Apply(applications []compass.Application) ([]Result, apperrors.AppError)
}

type service struct {
}

type Result struct {
	Operation Operation
	Error     apperrors.AppError
}

func (s *service) Apply(applications []compass.Application) ([]Result, apperrors.AppError) {
	return nil, nil
}
