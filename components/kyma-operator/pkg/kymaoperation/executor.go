package kymaoperation

import (
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/actionmanager"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	internalerrors "github.com/kyma-project/kyma/components/kyma-operator/pkg/errors"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymaoperation/actions"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	serviceCatalog "github.com/kyma-project/kyma/components/kyma-operator/pkg/servicecatalog"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/statusmanager"
)

// StepsProvider knows how to create an instance of the StepFactory
type StepsProvider interface {
	ForInstallation(*config.InstallationData, overrides.OverrideData) (actions.StepLister, error)
	ForUninstallation(*config.InstallationData) (actions.StepLister, error)
}

//Executor exposes top-level interfaxce for installation/uninstallation of Kyma
type Executor struct {
	serviceCatalog   serviceCatalog.ClientInterface
	errorHandlers    internalerrors.ErrorHandlersInterface
	statusManager    statusmanager.StatusManager
	actionManager    actionmanager.ActionManager
	stepsProvider    StepsProvider
	backoffIntervals []uint
}

// New .
func NewExecutor(serviceCatalog serviceCatalog.ClientInterface,
	statusManager statusmanager.StatusManager, actionManager actionmanager.ActionManager,
	stepsProvider StepsProvider, backoffIntervals []uint) *Executor {
	steps := &Executor{
		serviceCatalog:   serviceCatalog,
		errorHandlers:    &internalerrors.ErrorHandlers{},
		statusManager:    statusManager,
		actionManager:    actionManager,
		stepsProvider:    stepsProvider,
		backoffIntervals: backoffIntervals,
	}

	return steps
}

//InstallKyma .
func (steps *Executor) InstallKyma(installationData *config.InstallationData, overrideData overrides.OverrideData) error {

	removeLabelAndReturn := steps.removeLabelOnError(installationData.Context.Name, installationData.Context.Namespace)

	_ = steps.statusManager.InProgress("Verify installed components")

	installStepsProvider, err := steps.stepsProvider.ForInstallation(installationData, overrideData)
	if err != nil {
		_ = steps.statusManager.Error("Kyma Operator", "Verifying installed components", err)
		return err
	}

	stepList, err := installStepsProvider.StepList()

	if err != nil {
		return removeLabelAndReturn(err)
	}

	err = steps.processComponents(removeLabelAndReturn, installationData.Action, stepList)
	if err != nil {
		return err
	}

	err = steps.statusManager.InstallDone(installationData.URL, installationData.KymaVersion)
	if err != nil {
		return err
	}

	return nil
}

//UninstallKyma .
func (steps *Executor) UninstallKyma(installationData *config.InstallationData) error {

	removeLabelAndReturn := steps.removeLabelOnError(installationData.Context.Name, installationData.Context.Namespace)

	err := steps.DeprovisionAzureResources(DefaultDeprovisionConfig(), installationData.Context)
	steps.errorHandlers.LogError("An error during deprovisioning: ", err)

	_ = steps.statusManager.InProgress("Verify components to uninstall")

	uninstallStepsProvider, err := steps.stepsProvider.ForUninstallation(installationData)

	if err != nil {
		_ = steps.statusManager.Error("Kyma Operator", "Verify components to uninstall", err)
		return err
	}

	stepList, err := uninstallStepsProvider.StepList()

	if err != nil {
		return removeLabelAndReturn(err)
	}

	err = steps.processComponents(removeLabelAndReturn, installationData.Action, stepList)
	if steps.errorHandlers.CheckError("uninstall error: ", err) {
		return err
	}

	err = steps.statusManager.UninstallDone()
	if err != nil {
		return err
	}

	return nil
}

//PrintStep .
func (steps *Executor) PrintStep(stepName string) {
	log.Println("---------------------------")
	log.Println(stepName)
	log.Println("---------------------------")
}

func (steps *Executor) processComponents(removeLabelAndReturn func(err error) error, action string, stepList actions.StepList) error {

	log.Println("Processing Kyma components")

	logPrefix := action

	for _, step := range stepList {

		stepName := logPrefix + " component " + step.GetReleaseName()
		_ = steps.statusManager.InProgress(stepName)

		/*
		   //We prepare entire list of steps upfront
		   			step, err := stepsFactory.NewStep(component)
		   			if err != nil {
		   				return removeLabelAndReturn(err)
		   			}
		*/

		steps.PrintStep(stepName)

		err := retry.Do(
			func() error {
				processErr := step.Run()
				if steps.errorHandlers.CheckError("Step error: ", processErr) {
					_ = steps.statusManager.Error(step.GetReleaseName(), stepName, processErr)
					return processErr
				}
				return nil
			},
			retry.Attempts(uint(len(steps.backoffIntervals))+1),
			retry.DelayType(func(attempt uint, config *retry.Config) time.Duration {
				log.Printf("Warning: Retry number %d (sleeping for %d[s]).\n", attempt+1, steps.backoffIntervals[attempt])
				return time.Duration(steps.backoffIntervals[attempt]) * time.Second
			}),
		)

		if err != nil {
			return removeLabelAndReturn(fmt.Errorf("max number of retries reached during step: %s", stepName))
		}

		log.Println(stepName + "...DONE!")
	}

	log.Println(logPrefix + " Kyma components ...DONE!")

	return removeLabelAndReturn(nil)
}

func (steps *Executor) removeLabelOnError(installationCrName, installationCrNamespace string) func(err error) error {
	return func(err error) error {
		removeLabelError := steps.actionManager.RemoveActionLabel(installationCrName, installationCrNamespace)
		if steps.errorHandlers.CheckError("Error on removing label: ", removeLabelError) {
			err = fmt.Errorf("%v; Error on removing label: %v", err, removeLabelError)
		}
		return err
	}
}
