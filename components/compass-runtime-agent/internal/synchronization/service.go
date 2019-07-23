package synchronization

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
)

type Service struct {
}

type Result struct {
	Operation Operation
	Error     apperrors.AppError
}

func (s Service) Apply(applications []compass.Application) ([]Result, apperrors.AppError) {
	return nil, nil
}
