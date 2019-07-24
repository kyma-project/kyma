package synchronization

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
)

type reconciler struct {
}

type Operation int

const (
	Create Operation = iota
	Update
	Delete
)

type APIAction struct {
	Operation Operation
	API       graphql.APIDefinition
}

type EventAPIAction struct {
	Operation Operation
	EventAPI  graphql.EventAPIDefinition
}

type ApplicationAction struct {
	Operation       Operation
	Application     compass.Application
	APIActions      []APIAction
	EventAPIActions []EventAPIAction
}

func (r reconciler) Do(applications []compass.Application) ([]ApplicationAction, apperrors.AppError) {
	return nil, nil
}
