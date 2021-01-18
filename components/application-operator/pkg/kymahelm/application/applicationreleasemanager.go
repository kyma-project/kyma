package application

import (
	"context"
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	hapi_4 "helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	applicationChartDirectory = "application"
)

//go:generate mockery -name ApplicationClient
type ApplicationClient interface {
	List(context context.Context, opts v1.ListOptions) (*v1alpha1.ApplicationList, error)
	Update(context context.Context, app *v1alpha1.Application, opts v1.UpdateOptions) (*v1alpha1.Application, error)
}

//go:generate mockery -name ApplicationReleaseManager
type ApplicationReleaseManager interface {
	InstallChart(application *v1alpha1.Application) (hapi_4.Status, string, error)
	DeleteReleaseIfExists(name string) error
	CheckReleaseExistence(name string) (bool, error)
	CheckReleaseStatus(name string) (hapi_4.Status, string, error)
	UpgradeApplicationReleases() error
	UpgradeApplicationRelease(application *v1alpha1.Application)
}

type releaseManager struct {
	helmClient        kymahelm.HelmClient
	appClient         ApplicationClient
	overridesDefaults OverridesData
	namespace         string
	profile           string
}

func NewApplicationReleaseManager(helmClient kymahelm.HelmClient, appClient ApplicationClient, overridesDefaults OverridesData, namespace string, profile string) ApplicationReleaseManager {
	return &releaseManager{
		helmClient:        helmClient,
		appClient:         appClient,
		overridesDefaults: overridesDefaults,
		namespace:         namespace,
		profile:           profile,
	}
}

func (r *releaseManager) InstallChart(application *v1alpha1.Application) (hapi_4.Status, string, error) {
	overrides, err := r.prepareOverrides(application)
	if err != nil {
		return hapi_4.StatusFailed, "", errors.Wrapf(err, "Error parsing overrides for %s Application", application.Name)
	}

	installResponse, err := r.helmClient.InstallReleaseFromChart(applicationChartDirectory, application.Name, r.namespace, overrides, r.profile)
	if err != nil {
		return hapi_4.StatusFailed, "", err
	}
	return installResponse.Info.Status, installResponse.Info.Description, nil
}

func (r *releaseManager) UpgradeApplicationReleases() error {
	appList, err := r.appClient.List(context.Background(), v1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error fetching application list")
	}

	for _, app := range appList.Items {
		if app.ShouldSkipInstallation() == true {
			continue
		}

		r.UpgradeApplicationRelease(&app)
	}

	return nil
}

func (r *releaseManager) UpgradeApplicationRelease(app *v1alpha1.Application) {
	status, description, err := r.upgradeChart(app)
	if err != nil {
		log.Errorf("Failed to upgrade release %s: %s", app.Name, err.Error())

		setCurrentStatus(app, status.String(), description)

		err = r.updateApplication(app)
		if err != nil {
			log.Errorf("Failed to upgrade %s CR: %s", app.Name, err.Error())
		}
	}
}

func (r *releaseManager) upgradeChart(application *v1alpha1.Application) (hapi_4.Status, string, error) {

	releaseExist, err := r.checkExistence(application.Name)
	if err != nil {
		return hapi_4.StatusFailed, "", err
	}

	if releaseExist == false {
		return hapi_4.StatusUnknown, "", errors.Wrapf(err, "Application %s is missing", application.Name)
	}

	overrides, err := r.prepareOverrides(application)
	if err != nil {
		return hapi_4.StatusFailed, "", errors.Wrapf(err, "Error parsing overrides for %s Application", application.Name)
	}

	log.Infof("Upgrading release %s", application.Name)
	upgradeResponse, err := r.helmClient.UpdateReleaseFromChart(applicationChartDirectory, application.Name, r.namespace, overrides, r.profile)
	if err != nil {
		return hapi_4.StatusFailed, "", err
	}

	return upgradeResponse.Info.Status, upgradeResponse.Info.Description, nil
}

func (r *releaseManager) prepareOverrides(application *v1alpha1.Application) (map[string]interface{}, error) {
	overridesData := r.overridesDefaults

	if application.Spec.HasTenant() == true || application.Spec.HasGroup() == true {
		overridesData.Tenant = application.Spec.Tenant
		overridesData.Group = application.Spec.Group
	}

	var overridesMap map[string]interface{}
	bytes, err := json.Marshal(overridesData)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(bytes, &overridesMap); err != nil {
		return nil, err
	}

	overrides := map[string]interface{}{
		"global": overridesMap,
	}

	MergeLabelOverrides(application.Spec.Labels, overrides)

	return overrides, nil
}

func (r *releaseManager) DeleteReleaseIfExists(name string) error {
	releaseExist, err := r.checkExistence(name)
	if err != nil {
		return err
	}

	if releaseExist {
		_, err := r.helmClient.DeleteRelease(name, r.namespace)
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
	releases, err := r.helmClient.ListReleases(r.namespace)
	if err != nil {
		return false, err
	}

	for _, rel := range releases {
		if rel.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (r *releaseManager) CheckReleaseStatus(name string) (hapi_4.Status, string, error) {

	status, err := r.helmClient.ReleaseStatus(name, r.namespace)
	if err != nil {
		return hapi_4.StatusFailed, "", err
	}

	return status.Info.Status, status.Info.Description, nil
}

func (r *releaseManager) updateApplication(application *v1alpha1.Application) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, err := r.appClient.Update(context.Background(), application, v1.UpdateOptions{})
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
