package kymaoperation

import (
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/actionmanager"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	installationConfig "github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	internalerrors "github.com/kyma-project/kyma/components/kyma-operator/pkg/errors"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymaoperation/steps"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	serviceCatalog "github.com/kyma-project/kyma/components/kyma-operator/pkg/servicecatalog"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/statusmanager"
)

//StepsProvider knows how to create an instance of the StepFactory
type StepsProvider interface {
	ForInstallation(*config.InstallationData, overrides.OverrideData) (steps.StepLister, error)
	ForUninstallation(*config.InstallationData) (steps.StepLister, error)
}

//Executor exposes top-level interfaxce for installation/uninstallation of Kyma
type Executor struct {
	serviceCatalog              serviceCatalog.ClientInterface
	errorHandlers               internalerrors.ErrorHandlersInterface
	statusManager               statusmanager.StatusManager
	actionManager               actionmanager.ActionManager
	stepsProvider               StepsProvider
	deprovisionAzureResourcesFn func(extr *Executor, config *DeprovisionConfig, installation installationConfig.InstallationContext) error //TODO: Refactor!
	backoffIntervals            []uint
}

//NewExecutor creates a new instance of executor
func NewExecutor(serviceCatalog serviceCatalog.ClientInterface,
	statusManager statusmanager.StatusManager, actionManager actionmanager.ActionManager,
	stepsProvider StepsProvider, backoffIntervals []uint) *Executor {
	res := &Executor{
		serviceCatalog:              serviceCatalog,
		errorHandlers:               &internalerrors.ErrorHandlers{},
		statusManager:               statusManager,
		actionManager:               actionManager,
		stepsProvider:               stepsProvider,
		deprovisionAzureResourcesFn: DeprovisionAzureResources,
		backoffIntervals:            backoffIntervals,
	}

	return res
}

// InstallKyma is a top-level method used for Kyma installation
func (extr *Executor) InstallKyma(installationData *config.InstallationData, overrideData overrides.OverrideData) error {

	removeLabelAndReturn := extr.removeLabelOnError(installationData.Context.Name, installationData.Context.Namespace)

	_ = extr.statusManager.InProgress("Verify installed components")

	installStepsProvider, err := extr.stepsProvider.ForInstallation(installationData, overrideData)
	if err != nil {
		_ = extr.statusManager.Error("Kyma Operator", "Verifying installed components", err)
		return err
	}

	stepList, err := installStepsProvider.StepList()

	if err != nil {
		return removeLabelAndReturn(err)
	}

	err = extr.processComponents(removeLabelAndReturn, installationData.Action, stepList)
	if err != nil {
		return err
	}

	err = extr.statusManager.InstallDone(installationData.URL, installationData.KymaVersion)
	if err != nil {
		return err
	}

	return nil
}

//UninstallKyma is a top-level method used for Kyma uninstallation
func (extr *Executor) UninstallKyma(installationData *config.InstallationData) error {

	removeLabelAndReturn := extr.removeLabelOnError(installationData.Context.Name, installationData.Context.Namespace)

	err := extr.deprovisionAzureResourcesFn(extr, DefaultDeprovisionConfig(), installationData.Context)

	extr.errorHandlers.LogError("An error during deprovisioning: ", err)

	_ = extr.statusManager.InProgress("Verify components to uninstall")

	uninstallStepsProvider, err := extr.stepsProvider.ForUninstallation(installationData)

	if err != nil {
		_ = extr.statusManager.Error("Kyma Operator", "Verify components to uninstall", err)
		return err
	}

	stepList, err := uninstallStepsProvider.StepList()

	if err != nil {
		return removeLabelAndReturn(err)
	}

	err = extr.processComponents(removeLabelAndReturn, installationData.Action, stepList)
	if extr.errorHandlers.CheckError("uninstall error: ", err) {
		return err
	}

	err = extr.statusManager.UninstallDone()
	if err != nil {
		return err
	}

	return nil
}

//PrintStep .
func (extr *Executor) PrintStep(stepName string) {
	log.Println("---------------------------")
	log.Println(stepName)
	log.Println("---------------------------")
}

func (extr *Executor) processComponents(removeLabelAndReturn func(err error) error, action string, stepList steps.StepList) error {

	log.Println("Processing Kyma components")

	logPrefix := action

	for _, step := range stepList {

		stepName := logPrefix + " component " + step.GetReleaseName()
		_ = extr.statusManager.InProgress(stepName)

		extr.PrintStep(stepName)

		err := retry.Do(
			func() error {

				processErr := step.Run()
				if extr.errorHandlers.CheckError("Step error: ", processErr) {
					_ = extr.statusManager.Error(step.GetReleaseName(), stepName, processErr)
					return processErr
				}
				return nil
			},
			retry.Attempts(uint(len(extr.backoffIntervals))+1),
			retry.DelayType(func(attempt uint, config *retry.Config) time.Duration {
				log.Printf("Warning: Retry number %d (sleeping for %d[s]).\n", attempt+1, extr.backoffIntervals[attempt])
				return time.Duration(extr.backoffIntervals[attempt]) * time.Second
			}),
		)

		if err != nil {
			return removeLabelAndReturn(fmt.Errorf("max number of retries reached during step: %s", stepName))
		}

		log.Println(stepName + "... DONE!")
	}

	log.Println(logPrefix + " Kyma components ...DONE!")

	return removeLabelAndReturn(nil)
}

func (extr *Executor) removeLabelOnError(installationCrName, installationCrNamespace string) func(err error) error {
	return func(err error) error {
		removeLabelError := extr.actionManager.RemoveActionLabel(installationCrName, installationCrNamespace)
		if extr.errorHandlers.CheckError("Error on removing label: ", removeLabelError) {
			err = fmt.Errorf("%v; Error on removing label: %v", err, removeLabelError)
		}
		return err
	}
}
