package application_controller

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appReleases "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/application"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	installationSkippedStatus = "INSTALLATION_SKIPPED"
	applicationFinalizer      = "finalizer.applicationconnector.kyma-project.io"
)

type updateApplicationFunc func(application *v1alpha1.Application)

//go:generate mockery -name ApplicationManagerClient
type ApplicationManagerClient interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error
}

//go:generate mockery -name ApplicationReconciler
type ApplicationReconciler interface {
	Reconcile(request reconcile.Request) (reconcile.Result, error)
}

type applicationReconciler struct {
	applicationMgrClient ApplicationManagerClient
	releaseManager       appReleases.ApplicationReleaseManager
	log                  *logrus.Entry
}

func NewReconciler(appMgrClient ApplicationManagerClient, releaseManager appReleases.ApplicationReleaseManager, log *logrus.Entry) ApplicationReconciler {
	return &applicationReconciler{
		applicationMgrClient: appMgrClient,
		releaseManager:       releaseManager,
		log:                  log,
	}
}

func (r *applicationReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.Application{}

	r.log.Infof("Processing %s Application...", request.Name)

	err := r.applicationMgrClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return r.handleErrorWhileGettingInstance(err, request)
	}

	updateFunc, err := r.enforceDesiredState(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.updateApplicationCR(request.NamespacedName, updateFunc)
	if err != nil {
		return reconcile.Result{}, r.logAndError(err, "Error while updating Application %s", instance.Name)
	}

	return reconcile.Result{}, nil
}

func (r *applicationReconciler) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		r.log.Infof("Application %s deleted", request.Name)

		err := r.releaseManager.DeleteReleaseIfExists(request.Name)
		if err != nil {
			return reconcile.Result{}, r.logAndError(err, "")
		}
		r.log.Infof("Release %s successfully deleted", request.Name)

		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, r.logAndError(err, "Error getting %s Application", request.Name)
}

func (r *applicationReconciler) enforceDesiredState(application *v1alpha1.Application) (updateApplicationFunc, error) {
	if shouldBeRemoved(application) {
		return r.removeApplicationWithResources(application)
	}

	appStatus, statusDescription, err := r.manageInstallation(application)
	if err != nil {
		return nil, r.logAndError(err, "Error managing Helm release for %s Application", application.Name)
	}
	r.log.Infof("Release status for %s Application: %s", application.Name, appStatus)
	return func(application *v1alpha1.Application) {
		application.SetFinalizer(applicationFinalizer)
		application.SetAccessLabel()
		r.setCurrentStatus(application, appStatus, statusDescription)
	}, nil
}

func (r *applicationReconciler) removeApplicationWithResources(application *v1alpha1.Application) (updateApplicationFunc, error) {
	r.log.Infof("Removing %s Application with all resources...", application.Name)

	err := r.releaseManager.DeleteReleaseIfExists(application.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "Error removing %s Application with all resources", application.Name)
	}
	r.log.Infof("Release %s successfully deleted", application.Name)

	return func(application *v1alpha1.Application) {
		application.RemoveFinalizer(applicationFinalizer)
	}, nil
}

func (r *applicationReconciler) manageInstallation(application *v1alpha1.Application) (string, string, error) {
	releaseExist, err := r.releaseManager.CheckReleaseExistence(application.Name)
	if err != nil {
		return "", "", err
	}

	if !releaseExist {
		if application.ShouldSkipInstallation() == true {
			return installationSkippedStatus, "Installation will not be performed", nil
		}

		return r.installApplication(application)
	}

	return r.checkApplicationStatus(application)
}

func shouldBeRemoved(application *v1alpha1.Application) bool {
	return application.DeletionTimestamp != nil
}

func (r *applicationReconciler) installApplication(application *v1alpha1.Application) (string, string, error) {
	r.log.Infof("Installing release for %s Application...", application.Name)

	status, description, err := r.releaseManager.InstallChart(application)
	if err != nil {
		return "", "", errors.Wrapf(err, "Error installing release for %s Application", application.Name)
	}
	r.log.Infof("Release for %s Application installed successfully", application.Name)

	return status.String(), description, nil
}

func (r *applicationReconciler) checkApplicationStatus(application *v1alpha1.Application) (string, string, error) {
	status, description, err := r.releaseManager.CheckReleaseStatus(application.Name)
	if err != nil {
		return "", "", errors.Wrapf(err, "Error checking release status for %s Application", application.Name)
	}

	return status.String(), description, err
}

func (r *applicationReconciler) updateApplicationCR(namespacedName types.NamespacedName, updateFunc updateApplicationFunc) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instance := &v1alpha1.Application{}

		err := r.applicationMgrClient.Get(context.Background(), namespacedName, instance)
		if err != nil {
			return r.logAndError(err, "Error getting %s Application", namespacedName)
		}

		if instance.Spec.Services == nil {
			instance.Spec.Services = []v1alpha1.Service{}
		}

		updateFunc(instance)

		return r.applicationMgrClient.Update(context.Background(), instance)
	})
}

func (r *applicationReconciler) setCurrentStatus(application *v1alpha1.Application, status string, description string) {
	installationStatus := v1alpha1.InstallationStatus{
		Status:      status,
		Description: description,
	}

	application.SetInstallationStatus(installationStatus)
}

func (r *applicationReconciler) logAndError(err error, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	r.log.Errorf("%s: %s", msg, err.Error())
	return err
}
