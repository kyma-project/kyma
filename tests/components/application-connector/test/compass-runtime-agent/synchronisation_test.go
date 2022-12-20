package compass_runtime_agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	k8sErr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/executor"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/random"
)

const checkAppExistsPeriod = 10 * time.Second
const appCreationTimeout = 2 * time.Minute

type ApplicationReader interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Application, error)
}

func (cs *CompassRuntimeAgentSuite) TestApplication() {
	expectedAppName := "app1"
	compassAppName := expectedAppName + random.RandomString(10)

	//Create Application in Director
	applicationID, err := cs.directorClient.RegisterApplication(compassAppName, "Test Application for testing Compass Runtime Agent")
	cs.Require().NoError(err)

	synchronizedCompassAppName := fmt.Sprintf("mp-%s", compassAppName)

	applicationInterface := cs.applicationsClientSet.ApplicationconnectorV1alpha1().Applications()
	err = cs.assignApplicationToFormationAndWaitForSync(applicationInterface, synchronizedCompassAppName, applicationID)
	cs.Assert().NoError(err)

	// Compare Application created by Compass Runtime Agent with expected result

	cs.Run("Compass Runtime Agent should create Application", func() {
		err = cs.appComparator.Compare(cs.T(), expectedAppName, synchronizedCompassAppName)
		cs.Assert().NoError(err)
	})

	// Clean up
	cs.Run("Compass Runtime Agent should remove Application", func() {
		err = cs.removeApplicationAndWaitForSync(applicationInterface, synchronizedCompassAppName, applicationID)
		cs.NoError(err)
	})
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
		if err != nil {
			var statusErr *k8sErr.StatusError
			if errors.As(err, &statusErr) && statusErr.Status().Reason == v1.StatusReasonNotFound {
				t.Logf("Application was successfully removed by Compass Runtime Agent: %v", err)
				return true
			}
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
