package application

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"
)

const (
	applicationChartDirectory = "application"
)

type ApplicationClient interface {
	List(opts v1.ListOptions) (*v1alpha1.ApplicationList, error)
	Update(*v1alpha1.Application) (*v1alpha1.Application, error)
}

type ReleaseManager interface {
	InstallChart(application *v1alpha1.Application) (hapi_4.Status_Code, string, error)
	DeleteReleaseIfExists(name string) error
	CheckReleaseExistence(name string) (bool, error)
	CheckReleaseStatus(name string) (hapi_4.Status_Code, string, error)
	UpgradeReleases() error
}

type releaseManager struct {
	helmClient        kymahelm.HelmClient
	appClient         ApplicationClient
	overridesDefaults OverridesData
	namespace         string
}

func NewReleaseManager(helmClient kymahelm.HelmClient, appClient ApplicationClient, overridesDefaults OverridesData, namespace string) ReleaseManager {
	return &releaseManager{
		helmClient:        helmClient,
		appClient:         appClient,
		overridesDefaults: overridesDefaults,
		namespace:         namespace,
	}
}

func (r *releaseManager) InstallChart(application *v1alpha1.Application) (hapi_4.Status_Code, string, error) {
	overrides, err := r.prepareOverrides(application)
	if err != nil {
		return hapi_4.Status_FAILED, "", errors.Wrapf(err, "Error parsing overrides for %s Application", application.Name)
	}

	installResponse, err := r.helmClient.InstallReleaseFromChart(applicationChartDirectory, r.namespace, application.Name, overrides)
	if err != nil {
		return hapi_4.Status_FAILED, "", err
	}

	return installResponse.Release.Info.Status.Code, installResponse.Release.Info.Description, nil
}

func (r *releaseManager) UpgradeReleases() error {
	appList, err := r.appClient.List(v1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error fetching application list")
	}

	for _, app := range appList.Items {
		if app.ShouldSkipInstallation() == true {
			continue
		}

		status, description, err := r.upgradeChart(&app)
		if err != nil {
			log.Errorf("Failed to upgrade release %s: %s", app.Name, err.Error())

			setCurrentStatus(&app, status.String(), description)

			err = r.updateApplication(&app)
			if err != nil {
				log.Errorf("Failed to upgrade %s CR: %s", app.Name, err.Error())
			}
		}
	}

	return nil
}

func (r *releaseManager) upgradeChart(application *v1alpha1.Application) (hapi_4.Status_Code, string, error) {
	overrides, err := r.prepareOverrides(application)
	if err != nil {
		return hapi_4.Status_FAILED, "", errors.Wrapf(err, "Error parsing overrides for %s Application", application.Name)
	}

	log.Infof("Upgrading release %s", application.Name)
	upgradeResponse, err := r.helmClient.UpdateReleaseFromChart(applicationChartDirectory, application.Name, overrides)
	if err != nil {
		return hapi_4.Status_FAILED, "", err
	}

	return upgradeResponse.Release.Info.Status.Code, upgradeResponse.Release.Info.Description, nil
}

func (r *releaseManager) prepareOverrides(application *v1alpha1.Application) (string, error) {
	overridesData := r.overridesDefaults

	if application.Spec.HasTenant() == true || application.Spec.HasGroup() == true {
		overridesData.Tenant = application.Spec.Tenant
		overridesData.Group = application.Spec.Group
	}

	return kymahelm.ParseOverrides(overridesData, overridesTemplate)
}

func (r *releaseManager) DeleteReleaseIfExists(name string) error {
	releaseExist, err := r.checkExistence(name)
	if err != nil {
		return err
	}

	if releaseExist {
		_, err := r.helmClient.DeleteRelease(name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *releaseManager) CheckReleaseExistence(name string) (bool, error) {
	return r.checkExistence(name)
}

func (r *releaseManager) checkExistence(name string) (bool, error) {
	listResponse, err := r.helmClient.ListReleases(r.namespace)
	if err != nil {
		return false, err
	}

	releases := listResponse.Releases

	for _, rel := range releases {
		if rel.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (r *releaseManager) CheckReleaseStatus(name string) (hapi_4.Status_Code, string, error) {
	status, err := r.helmClient.ReleaseStatus(name)
	if err != nil {
		return hapi_4.Status_FAILED, "", err
	}

	return status.Info.Status.Code, status.Info.Description, nil
}

func (r *releaseManager) updateApplication(application *v1alpha1.Application) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, err := r.appClient.Update(application)
		return err
	})
}

func setCurrentStatus(application *v1alpha1.Application, status string, description string) {
	installationStatus := v1alpha1.InstallationStatus{
		Status:      status,
		Description: description,
	}

	application.SetInstallationStatus(installationStatus)
}
