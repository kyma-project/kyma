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
	UpdateDone(url, kymaVersion string) error
	UninstallDone() error
	Error(description string) error
}

type statusManager struct {
	lister listers.InstallationLister
	client installationClientset.Interface
}

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
	instObj, getErr := sm.lister.Installations(consts.InstNamespace).Get(consts.InstResource)
	if getErr != nil {
		return getErr
	}

	installationCopy := instObj.DeepCopy()

	instStatus := getStatus(installationv1alpha1.StateInProgress, description, installationCopy.Status.URL, installationCopy.Status.KymaVersion)
	return sm.update(instStatus)
}

//InstallDone .
func (sm *statusManager) InstallDone(url, kymaVersion string) error {
	instStatus := getStatus(installationv1alpha1.StateInstalled, "Kyma installed", url, kymaVersion)
	return sm.update(instStatus)
}

//UpdateDone .
func (sm *statusManager) UpdateDone(url, kymaVersion string) error {
	instStatus := getStatus(installationv1alpha1.StateUpdated, "Kyma updated", url, kymaVersion)
	return sm.update(instStatus)
}

//UninstallDone .
func (sm *statusManager) UninstallDone() error {
	instStatus := getStatus(installationv1alpha1.StateUninstalled, "Kyma uninstalled", "", "")
	return sm.update(instStatus)
}

//Error .
func (sm *statusManager) Error(description string) error {
	instStatus := getStatus(installationv1alpha1.StateError, description, "", "")
	return sm.update(instStatus)
}

func (sm *statusManager) update(installationStatus *installationv1alpha1.InstallationStatus) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instObj, getErr := sm.lister.Installations(consts.InstNamespace).Get(consts.InstResource)
		if getErr != nil {
			return getErr
		}

		installationCopy := instObj.DeepCopy()
		installationCopy.Status = *installationStatus

		installationCopy.Status.Conditions = instObj.Status.Conditions

		_, updateErr := sm.client.InstallerV1alpha1().Installations(consts.InstNamespace).Update(installationCopy)
		return updateErr
	})

	if retryErr != nil {
		return retryErr
	}

	return nil
}

func getStatus(state installationv1alpha1.StateEnum, description, url, version string) *installationv1alpha1.InstallationStatus {
	return &installationv1alpha1.InstallationStatus{
		State:       state,
		Description: description,
		URL:         url,
		KymaVersion: version,
	}
}
