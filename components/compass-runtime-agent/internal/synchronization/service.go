package synchronization

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
)

type Service struct {
}

type ApplicationEntry struct {
	Application        graphql.Application
	APIDefinition      graphql.APIDefinition
	EventAPIDefinition graphql.EventAPIDefinition
}

type Result struct {
	Operation Operation
	Error     apperrors.AppError
}

func (s Service) Apply(applications []ApplicationEntry) (apperrors.AppError, []Result) {
	return nil, nil
}
