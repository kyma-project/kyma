package synchronization

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
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
	API       APIDefinition
}

type EventAPIAction struct {
	Operation Operation
	EventAPI  EventAPIDefinition
}

type ApplicationAction struct {
	Operation       Operation
	Application     Application
	APIActions      []APIAction
	EventAPIActions []EventAPIAction
}

func (r reconciler) Do(applications []Application) ([]ApplicationAction, apperrors.AppError) {
	return nil, nil
}
