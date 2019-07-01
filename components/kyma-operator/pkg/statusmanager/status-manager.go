package statusmanager

import (
	installationv1alpha1 "github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	installationClientset "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned"
	listers "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/listers/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/consts"
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
		status.Description = "Selected Kyma components uninstalled"
	})
}

//Error .
func (sm *statusManager) Error(component, description string, err error) error {
	return sm.updateFunc(func(status *installationv1alpha1.InstallationStatus) {
		status.State = installationv1alpha1.StateError
		status.Description = description
		status.URL = ""
		status.KymaVersion = ""
		logEntry := installationv1alpha1.ErrorLogEntry{
			Component:   component,
			Log:         err.Error(),
			Occurrences: 1,
		}
		status.ErrorLog = appendErrorLog(status.ErrorLog, logEntry)
	})
}

func appendErrorLog(entries []installationv1alpha1.ErrorLogEntry, newEntry installationv1alpha1.ErrorLogEntry) []installationv1alpha1.ErrorLogEntry {
	if len(entries) == 0 {
		return []installationv1alpha1.ErrorLogEntry{newEntry}
	}

	lastEntry := entries[len(entries)-1]
	if lastEntry.Component == newEntry.Component && lastEntry.Log == newEntry.Log {
		lastEntry.Occurrences += 1
		entries[len(entries)-1] = lastEntry
		return entries
	}

	return append(entries, newEntry)
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
