package sync

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type reconciler struct {
	applicationsInterface applications.Manager
	converter             applications.Converter
}

type Operation int

const (
	Create Operation = iota
	Update
	Delete
)

//go:generate mockery -name=Reconciler
type Reconciler interface {
	Do(applications []v1alpha1.Application) ([]ApplicationAction, apperrors.AppError)
}

type ServiceAction struct {
	Operation Operation
	Service   v1alpha1.Service
}

type EventAPIAction struct {
	Operation Operation
	EventAPI  v1alpha1.Service
}

type ApplicationAction struct {
	Operation      Operation
	Application    v1alpha1.Application
	ServiceActions []ServiceAction
}

func NewReconciler(applicationsInterface applications.Manager) Reconciler {
	return reconciler{
		applicationsInterface: applicationsInterface,
	}
}

func (r reconciler) Do(directorApplications []v1alpha1.Application) ([]ApplicationAction, apperrors.AppError) {

	actions := make([]ApplicationAction, 0, len(directorApplications))
	existingApplications, err := r.applicationsInterface.List(v1.ListOptions{})
	if err != nil {
		apperrors.Internal("Failed to get applications list: %s", err)
	}

	new := r.getNewApps(directorApplications, existingApplications)
	deleted := r.getDeleted(directorApplications, existingApplications)
	updated := r.getUpdated(directorApplications, existingApplications)

	actions = append(actions, new...)
	actions = append(actions, deleted...)
	actions = append(actions, updated...)

	return actions, nil
}

func (r reconciler) getNewApps(directorApplications []v1alpha1.Application, runtimeApplications *v1alpha1.ApplicationList) []ApplicationAction {

	actions := make([]ApplicationAction, 0)

	for _, application := range directorApplications {
		found := applications.ApplicationExists(application.Name, runtimeApplications)
		if !found {
			actions = append(actions, ApplicationAction{
				Operation:      Create,
				Application:    application,
				ServiceActions: r.getNewServices(application.Spec.Services, v1alpha1.Application{}),
			})
		}
	}

	return actions
}

func (r reconciler) getNewServices(services []v1alpha1.Service, application v1alpha1.Application) []ServiceAction {

	actions := make([]ServiceAction, 0)

	for _, service := range services {
		found := applications.ServiceExists(service.ID, application)
		if !found {
			actions = append(actions, ServiceAction{
				Operation: Create,
				Service:   service,
			})
		}
	}

	return actions
}

func (r reconciler) getDeleted(directorApplications []v1alpha1.Application, runtimeApplications *v1alpha1.ApplicationList) []ApplicationAction {
	return nil
}

func (r reconciler) getUpdated(directorApplications []v1alpha1.Application, runtimeApplications *v1alpha1.ApplicationList) []ApplicationAction {
	return nil
}
