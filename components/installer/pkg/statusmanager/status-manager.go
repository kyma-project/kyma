package statusmanager

import (
	installationv1alpha1 "github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	installationClientset "github.com/kyma-project/kyma/components/installer/pkg/client/clientset/versioned"
	listers "github.com/kyma-project/kyma/components/installer/pkg/client/listers/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"k8s.io/client-go/util/retry"
)

// StatusManager .
type StatusManager interface {
	InProgress(description string) error
	InstallDone(url, kymaVersion string) error
	UninstallDone() error
	Error(component, description string, err error) error
}

type statusManager struct {
	lister listers.InstallationLister
	client installationClientset.Interface
}

type statusUpdater func(*installationv1alpha1.InstallationStatus)

// NewKymaStatusManager .
func NewKymaStatusManager(installationClient installationClientset.Interface, installationLister listers.InstallationLister) StatusManager {
	sm := &statusManager{
		lister: installationLister,
		client: installationClient,
	}

	return sm
}

//InProgress .
func (sm *statusManager) InProgress(description string) error {
	return sm.updateFunc(func(status *installationv1alpha1.InstallationStatus) {
		status.State = installationv1alpha1.StateInProgress
		status.Description = description
	})
}

//InstallDone .
func (sm *statusManager) InstallDone(url, kymaVersion string) error {
	return sm.updateFunc(func(status *installationv1alpha1.InstallationStatus) {
		status.State = installationv1alpha1.StateInstalled
		status.Description = "Kyma installed"
		status.URL = url
		status.KymaVersion = kymaVersion
		status.ErrorLog = []installationv1alpha1.ErrorLogEntry{}
	})
}

//UninstallDone .
func (sm *statusManager) UninstallDone() error {
	return sm.updateFunc(func(status *installationv1alpha1.InstallationStatus) {
		status.State = installationv1alpha1.StateUninstalled
		status.Description = "Kyma uninstalled"
	})
}

//Error .
func (sm *statusManager) Error(component, description string, err error) error {
	return sm.updateFunc(func(status *installationv1alpha1.InstallationStatus) {
		status.State = installationv1alpha1.StateError
		status.Description = description
		status.URL = ""
		status.KymaVersion = ""
		logEntry := installationv1alpha1.ErrorLogEntry{Component: component, Log: err.Error()}
		status.ErrorLog = append(status.ErrorLog, logEntry)
	})
}

func (sm *statusManager) updateFunc(updater statusUpdater) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instObj, getErr := sm.lister.Installations(consts.InstNamespace).Get(consts.InstResource)
		if getErr != nil {
			return getErr
		}

		installationCopy := instObj.DeepCopy()
		updater(&installationCopy.Status)

		_, updateErr := sm.client.InstallerV1alpha1().Installations(consts.InstNamespace).Update(installationCopy)
		return updateErr
	})

	if retryErr != nil {
		return retryErr
	}

	return nil
}
