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

	new := r.getNew(directorApplications, existingApplications)
	deleted := r.getDeleted(directorApplications, existingApplications)
	updated := r.getUpdated(directorApplications, existingApplications)

	actions = append(actions, new...)
	actions = append(actions, deleted...)
	actions = append(actions, updated...)

	return actions, nil
}

func (r reconciler) getNew(directorApplications []model.Application, runtimeApplications *v1alpha1.ApplicationList) []ApplicationAction {

	actions := make([]ApplicationAction, 0)

	find := func(applicationName string) bool {
		if runtimeApplications == nil {
			return false
		}

		for _, runtimeApplication := range runtimeApplications.Items {
			if runtimeApplication.Name == applicationName {
				return true
			}
		}

		return false
	}

	for _, directorApp := range directorApplications {
		found := find(directorApp.Name)
		if found {
			actions = append(actions, ApplicationAction{
				Operation:   Create,
				Application: directorApp,
			})
		}
	}

	return nil
}

func (r reconciler) getDeleted(directorApplications []model.Application, runtimeApplications *v1alpha1.ApplicationList) []ApplicationAction {
	return nil
}

func (r reconciler) getUpdated(directorApplications []model.Application, runtimeApplications *v1alpha1.ApplicationList) []ApplicationAction {
	return nil
}
