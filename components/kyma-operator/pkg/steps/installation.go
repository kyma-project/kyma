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
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
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

	legacyKymaSourceConfig := kymasources.LegacyKymaSourceConfig{
		KymaURL:     installationData.URL,
		KymaVersion: installationData.KymaVersion,
	}

	stepsFactory, factoryErr := steps.stepFactoryCreator.NewInstallStepFactory(overrideData, legacyKymaSourceConfig)
	if factoryErr != nil {
		_ = steps.statusManager.Error("Kyma Operator", "Verify installed components", factoryErr)
		return factoryErr
	}

	err := steps.processComponents(installationData, stepsFactory)
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

	stepsFactory, factoryErr := steps.stepFactoryCreator.NewUninstallStepFactory()
	if factoryErr != nil {
		_ = steps.statusManager.Error("Kyma Operator", "Verify components to uninstall", factoryErr)
		return factoryErr
	}

	err = steps.processComponents(installationData, stepsFactory)
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

func (steps *InstallationSteps) processComponents(installationData *config.InstallationData, stepsFactory kymainstallation.StepFactory) error {

	log.Println("Processing Kyma components")

	removeLabelAndReturn := func(err error) error {
		removeLabelError := steps.actionManager.RemoveActionLabel(installationData.Context.Name, installationData.Context.Namespace, "action")
		if steps.errorHandlers.CheckError("Error on removing label: ", err) {
			err = fmt.Errorf("%v; Error on removing label: %v", err, removeLabelError)
		}
		return err
	}

	logPrefix := installationData.Action

	for _, component := range installationData.Components {

		stepName := logPrefix + " component " + component.GetReleaseName()
		_ = steps.statusManager.InProgress(stepName)

		step := stepsFactory.NewStep(component)

		steps.PrintStep(stepName)

		err := retry.Do(
			func() error {
				processErr := step.Run()
				if steps.errorHandlers.CheckError("Step error: ", processErr) {
					_ = steps.statusManager.Error(component.GetReleaseName(), stepName, processErr)
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
			return removeLabelAndReturn(fmt.Errorf("Max number of retries reached during step: %s", stepName))
		}

		log.Println(stepName + "...DONE!")
	}

	log.Println(logPrefix + " Kyma components ...DONE!")

	return removeLabelAndReturn(nil)
}
