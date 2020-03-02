package conditionmanager

import (
	installationv1alpha1 "github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	installationClientset "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned"
	listers "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/listers/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/consts"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// Interface .
type Interface interface {
	InstallStart() error
	InstallSuccess() error
	InstallError() error

	UninstallStart() error
	UninstallSuccess() error
	UninstallError() error
}

type impl struct {
	lister listers.InstallationLister
	client installationClientset.Interface
}

func (cm *impl) InstallStart() error {
	installation, err := cm.getInstallation()
	if err != nil {
		return err
	}

	cm.setCondition(installation, installationv1alpha1.CondtitionInstalled, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionInstalling, v1.ConditionTrue)
	cm.setCondition(installation, installationv1alpha1.ConditionInProgress, v1.ConditionTrue)
	cm.setCondition(installation, installationv1alpha1.ConditionError, v1.ConditionFalse)

	err = cm.update(installation)
	if err != nil {
		return err
	}

	return nil
}

func (cm *impl) InstallSuccess() error {
	installation, err := cm.getInstallation()
	if err != nil {
		return err
	}

	cm.setCondition(installation, installationv1alpha1.CondtitionInstalled, v1.ConditionTrue)
	cm.setCondition(installation, installationv1alpha1.ConditionInstalling, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionInProgress, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionUninstalled, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionError, v1.ConditionFalse)

	err = cm.update(installation)
	if err != nil {
		return err
	}

	return nil
}

func (cm *impl) InstallError() error {
	installation, err := cm.getInstallation()
	if err != nil {
		return err
	}

	cm.setCondition(installation, installationv1alpha1.CondtitionInstalled, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionInstalling, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionInProgress, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionError, v1.ConditionTrue)

	err = cm.update(installation)
	if err != nil {
		return err
	}

	return nil
}

func (cm *impl) UninstallStart() error {
	installation, err := cm.getInstallation()
	if err != nil {
		return err
	}

	cm.setCondition(installation, installationv1alpha1.ConditionUninstalling, v1.ConditionTrue)
	cm.setCondition(installation, installationv1alpha1.ConditionInProgress, v1.ConditionTrue)
	cm.setCondition(installation, installationv1alpha1.ConditionError, v1.ConditionFalse)

	err = cm.update(installation)
	if err != nil {
		return err
	}

	return nil
}

func (cm *impl) UninstallSuccess() error {
	installation, err := cm.getInstallation()
	if err != nil {
		return err
	}

	cm.setCondition(installation, installationv1alpha1.ConditionUninstalling, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionUninstalled, v1.ConditionTrue)
	cm.setCondition(installation, installationv1alpha1.CondtitionInstalled, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionInstalling, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionInProgress, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionError, v1.ConditionFalse)

	err = cm.update(installation)
	if err != nil {
		return err
	}

	return nil
}

func (cm *impl) UninstallError() error {
	installation, err := cm.getInstallation()
	if err != nil {
		return err
	}

	cm.setCondition(installation, installationv1alpha1.ConditionUninstalling, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionUninstalled, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionInProgress, v1.ConditionFalse)
	cm.setCondition(installation, installationv1alpha1.ConditionError, v1.ConditionTrue)

	err = cm.update(installation)
	if err != nil {
		return err
	}

	return nil
}

func (cm *impl) getInstallation() (*installationv1alpha1.Installation, error) {
	installation, err := cm.lister.Installations(consts.InstNamespace).Get(consts.InstResource)
	if err != nil {
		return nil, err
	}

	return installation.DeepCopy(), nil
}

func (cm *impl) getOrCreateCondition(installation *installationv1alpha1.Installation, conditionType installationv1alpha1.InstallationConditionType) *installationv1alpha1.InstallationCondition {
	condition := cm.getCondition(installation, conditionType)
	if condition != nil {
		return condition
	}
	cm.createCondition(installation, conditionType)
	return cm.getOrCreateCondition(installation, conditionType)
}

func (cm *impl) getCondition(installation *installationv1alpha1.Installation, conditionType installationv1alpha1.InstallationConditionType) *installationv1alpha1.InstallationCondition {
	for i := 0; i < len(installation.Status.Conditions); i++ {
		if installation.Status.Conditions[i].Type != conditionType {
			continue
		}

		return &installation.Status.Conditions[i]
	}

	return nil
}

func (cm *impl) createCondition(installation *installationv1alpha1.Installation, conditionType installationv1alpha1.InstallationConditionType) {
	condition := &installationv1alpha1.InstallationCondition{
		Type:   conditionType,
		Status: v1.ConditionUnknown,
	}

	installation.Status.Conditions = append(installation.Status.Conditions, *condition)
}

func (cm *impl) setCondition(installation *installationv1alpha1.Installation, conditionType installationv1alpha1.InstallationConditionType, status v1.ConditionStatus) {
	condition := cm.getOrCreateCondition(installation, conditionType)
	if condition.Status != status {
		condition.Status = status
		condition.LastTransitionTime = metav1.Now()
	}
	condition.LastProbeTime = metav1.Now()
}

func (cm *impl) update(installation *installationv1alpha1.Installation) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instObj, getErr := cm.lister.Installations(consts.InstNamespace).Get(consts.InstResource)
		if getErr != nil {
			return getErr
		}

		installationCopy := instObj.DeepCopy()
		installationCopy.Status.Conditions = installation.Status.Conditions

		_, updateErr := cm.client.InstallerV1alpha1().Installations(consts.InstNamespace).Update(installationCopy)

		return updateErr
	})

	if retryErr != nil {
		return retryErr
	}

	return nil
}

// New .
func New(installationClient installationClientset.Interface, installationLister listers.InstallationLister) Interface {
	return &impl{
		client: installationClient,
		lister: installationLister,
	}
}
