package steps

import (
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/actionmanager"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	internalerrors "github.com/kyma-project/kyma/components/kyma-operator/pkg/errors"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymainstallation"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	serviceCatalog "github.com/kyma-project/kyma/components/kyma-operator/pkg/servicecatalog"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/statusmanager"
)

//InstallationSteps .
type InstallationSteps struct {
	serviceCatalog     serviceCatalog.ClientInterface
	errorHandlers      internalerrors.ErrorHandlersInterface
	statusManager      statusmanager.StatusManager
	actionManager      actionmanager.ActionManager
	stepFactoryCreator kymainstallation.StepFactoryCreator
	backoffIntervals   []uint
}

// New .
func New(serviceCatalog serviceCatalog.ClientInterface,
	statusManager statusmanager.StatusManager, actionManager actionmanager.ActionManager,
	stepFactoryCreator kymainstallation.StepFactoryCreator, backoffIntervals []uint) *InstallationSteps {
	steps := &InstallationSteps{
		serviceCatalog:     serviceCatalog,
		errorHandlers:      &internalerrors.ErrorHandlers{},
		statusManager:      statusManager,
		actionManager:      actionManager,
		stepFactoryCreator: stepFactoryCreator,
		backoffIntervals:   backoffIntervals,
	}

	return steps
}

//InstallKyma .
func (steps *InstallationSteps) InstallKyma(installationData *config.InstallationData, overrideData overrides.OverrideData) error {

	_ = steps.statusManager.InProgress("Verify installed components")

	stepsFactory, factoryErr := steps.stepFactoryCreator.NewInstallStepFactory(installationData, overrideData)
	if factoryErr != nil {
		_ = steps.statusManager.Error("Kyma Operator", "Verify installed components", factoryErr)
		return factoryErr
	}

	removeLabelAndReturn := steps.removeLabelOnError(installationData.Context.Name, installationData.Context.Namespace)

	stepList, err := stepsFactory.InstallStepList()

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
func (steps *InstallationSteps) UninstallKyma(installationData *config.InstallationData) error {

	err := steps.DeprovisionAzureResources(DefaultDeprovisionConfig(), installationData.Context)
	steps.errorHandlers.LogError("An error during deprovisioning: ", err)

	_ = steps.statusManager.InProgress("Verify components to uninstall")

	stepsFactory, factoryErr := steps.stepFactoryCreator.NewUninstallStepFactory(installationData)
	if factoryErr != nil {
		_ = steps.statusManager.Error("Kyma Operator", "Verify components to uninstall", factoryErr)
		return factoryErr
	}

	removeLabelAndReturn := steps.removeLabelOnError(installationData.Context.Name, installationData.Context.Namespace)

	stepList, err := stepsFactory.UninstallStepList()
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
func (steps *InstallationSteps) PrintStep(stepName string) {
	log.Println("---------------------------")
	log.Println(stepName)
	log.Println("---------------------------")
}

func (steps *InstallationSteps) processComponents(removeLabelAndReturn func(err error) error, action string, stepList kymainstallation.StepList) error {

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

func (steps *InstallationSteps) removeLabelOnError(installationCrName, installationCrNamespace string) func(err error) error {
	return func(err error) error {
		removeLabelError := steps.actionManager.RemoveActionLabel(installationCrName, installationCrNamespace)
		if steps.errorHandlers.CheckError("Error on removing label: ", removeLabelError) {
			err = fmt.Errorf("%v; Error on removing label: %v", err, removeLabelError)
		}
		return err
	}
}
