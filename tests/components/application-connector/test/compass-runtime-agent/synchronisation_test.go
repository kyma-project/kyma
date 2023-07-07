package compass_runtime_agent

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/executor"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/random"
)

const checkAppExistsPeriod = 10 * time.Second
const appCreationTimeout = 2 * time.Minute
const appUpdateTimeout = 2 * time.Minute

const updatedDescription = "The app was updated"

type ApplicationReader interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Application, error)
}

func (cs *CompassRuntimeAgentSuite) TestApplication() {
	expectedAppName := "app1"
	updatedAppName := "app1-updated"

	compassAppName := expectedAppName + random.RandomString(10)

	correctState := false

	//Create Application in Director
	applicationID, err := cs.directorClient.RegisterApplication(compassAppName, "Test Application for testing Compass Runtime Agent")
	cs.Require().NoError(err)

	synchronizedCompassAppName := fmt.Sprintf("mp-%s", compassAppName)

	applicationInterface := cs.applicationsClientSet.ApplicationconnectorV1alpha1().Applications()
	err = cs.assignApplicationToFormationAndWaitForSync(applicationInterface, synchronizedCompassAppName, applicationID)
	cs.NoError(err)

	// Compare Application created by Compass Runtime Agent with expected result
	cs.Run("Compass Runtime Agent should create Application", func() {
		err = cs.appComparator.Compare(cs.T(), expectedAppName, synchronizedCompassAppName)
		cs.NoError(err)

		correctState = err == nil
	})

	cs.Run("Update app", func() {
		if !correctState {
			cs.T().Skip("App not in correct state")
		}

		_ = cs.updateAndWait(applicationInterface, synchronizedCompassAppName, applicationID)

		err = cs.appComparator.Compare(cs.T(), updatedAppName, synchronizedCompassAppName)
		cs.NoError(err)

		correctState = err == nil
	})

	// Clean up
	cs.Run("Compass Runtime Agent should remove Application", func() {
		err = cs.removeApplicationAndWaitForSync(applicationInterface, synchronizedCompassAppName, applicationID)
		cs.NoError(err)
	})
}

func (cs *CompassRuntimeAgentSuite) updateAndWait(appReader ApplicationReader, compassAppName, applicationID string) error {
	t := cs.T()
	t.Helper()

	exec := func() error {
		_, err := cs.directorClient.UpdateApplication(applicationID, updatedDescription)
		return err
	}

	verify := func() bool {
		app, err := appReader.Get(context.Background(), compassAppName, v1.GetOptions{})
		if err != nil {
			t.Logf("Couldn't get updated: %v", err)
		}

		return err == nil && app.Spec.Description == updatedDescription
	}

	return executor.ExecuteAndWaitForCondition{
		RetryableExecuteFunc: exec,
		ConditionMetFunc:     verify,
		Tick:                 checkAppExistsPeriod,
		Timeout:              appUpdateTimeout,
	}.Do()
}

func (cs *CompassRuntimeAgentSuite) assignApplicationToFormationAndWaitForSync(appReader ApplicationReader, compassAppName, applicationID string) error {
	t := cs.T()
	t.Helper()

	exec := func() error {
		return cs.directorClient.AssignApplicationToFormation(applicationID, cs.formationName)
	}

	verify := func() bool {
		_, err := appReader.Get(context.Background(), compassAppName, v1.GetOptions{})
		if err != nil {
			t.Logf("Failed to get app: %v", err)
		}

		return err == nil
	}

	return executor.ExecuteAndWaitForCondition{
		RetryableExecuteFunc: exec,
		ConditionMetFunc:     verify,
		Tick:                 checkAppExistsPeriod,
		Timeout:              appCreationTimeout,
	}.Do()
}

func (cs *CompassRuntimeAgentSuite) removeApplicationAndWaitForSync(appReader ApplicationReader, compassAppName, applicationID string) error {
	t := cs.T()
	t.Helper()

	exec := func() error {
		err := cs.directorClient.UnassignApplication(applicationID, cs.formationName)
		if err != nil {
			return err
		}

		err = cs.directorClient.UnregisterApplication(applicationID)
		return err
	}

	verify := func() bool {
		_, err := appReader.Get(context.Background(), compassAppName, v1.GetOptions{})
		if errors.IsNotFound(err) {
			t.Logf("Application was successfully removed by Compass Runtime Agent: %v", err)
			return true
		}

		if err != nil {
			t.Logf("Failed to check whether Application was removed by Compass Runtime Agent: %v", err)
		}

		return false
	}

	return executor.ExecuteAndWaitForCondition{
		RetryableExecuteFunc: exec,
		ConditionMetFunc:     verify,
		Tick:                 checkAppExistsPeriod,
		Timeout:              appCreationTimeout,
	}.Do()
}
