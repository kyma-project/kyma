package sync

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type reconciler struct {
	applicationsInterface applications.Manager
}

type Operation int

const (
	Create Operation = iota
	Update
	Delete
)

//go:generate mockery -name=Reconciler
type Reconciler interface {
	Do(applications []model.Application) ([]ApplicationAction, apperrors.AppError)
}

type APIAction struct {
	Operation Operation
	API       model.APIDefinition
}

type EventAPIAction struct {
	Operation Operation
	EventAPI  model.EventAPIDefinition
}

// TODO: consider using v1alpha1.Application type here
type ApplicationAction struct {
	Operation       Operation
	Application     model.Application
	APIActions      []APIAction
	EventAPIActions []EventAPIAction
}

func NewReconciler(applicationsInterface applications.Manager) Reconciler {
	return reconciler{
		applicationsInterface: applicationsInterface,
	}
}

func (r reconciler) Do(directorApplications []model.Application) ([]ApplicationAction, apperrors.AppError) {

	actions := make([]ApplicationAction, 0, len(directorApplications))
	existingApplications, err := r.applicationsInterface.List(v1.ListOptions{})
	if err != nil {
		apperrors.Internal("Failed tp get applications list: %s", err)
	}

	new := r.getNewApps(directorApplications, existingApplications)
	deleted := r.getDeleted(directorApplications, existingApplications)
	updated := r.getUpdated(directorApplications, existingApplications)

	actions = append(actions, new...)
	actions = append(actions, deleted...)
	actions = append(actions, updated...)

	return actions, nil
}

func (r reconciler) getNewApps(directorApplications []model.Application, runtimeApplications *v1alpha1.ApplicationList) []ApplicationAction {

	actions := make([]ApplicationAction, 0)

	for _, directorApp := range directorApplications {
		found := applications.ApplicationExists(directorApp.Name, runtimeApplications)
		if !found {
			actions = append(actions, ApplicationAction{
				Operation:       Create,
				Application:     directorApp,
				APIActions:      r.getNewApis(directorApp.APIs, v1alpha1.Application{}),
				EventAPIActions: r.getNewEventApis(directorApp.EventAPIs, v1alpha1.Application{}),
			})
		}
	}

	return actions
}

func (r reconciler) getNewApis(directorAPIs []model.APIDefinition, application v1alpha1.Application) []APIAction {

	actions := make([]APIAction, 0)

	for _, directorAPI := range directorAPIs {
		found := applications.ServiceExists(directorAPI.ID, application)
		if !found {
			actions = append(actions, APIAction{
				Operation: Create,
				API:       directorAPI,
			})
		}
	}

	return actions
}

func (r reconciler) getNewEventApis(directorEventAPIs []model.EventAPIDefinition, application v1alpha1.Application) []EventAPIAction {
	actions := make([]EventAPIAction, 0)

	for _, directorEventAPI := range directorEventAPIs {
		found := applications.ServiceExists(directorEventAPI.ID, application)
		if !found {
			actions = append(actions, EventAPIAction{
				Operation: Create,
				EventAPI:  directorEventAPI,
			})
		}
	}

	return actions
}

func (r reconciler) getDeleted(directorApplications []model.Application, runtimeApplications *v1alpha1.ApplicationList) []ApplicationAction {
	return nil
}

func (r reconciler) getUpdated(directorApplications []model.Application, runtimeApplications *v1alpha1.ApplicationList) []ApplicationAction {
	return nil
}
