package compass_runtime_agent

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/executor"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/random"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const checkAppExistsPeriod = 10 * time.Second
const appCreationTimeout = 2 * time.Minute

type ApplicationReader interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Application, error)
}

func (gs *CompassRuntimeAgentSuite) TestCreatingApplications() {
	expectedAppName := "app1"
	compassAppName := expectedAppName + random.RandomString(10)

	//Create Application in Director and wait until it gets created
	applicationID, err := gs.directorClient.RegisterApplication(compassAppName, "Test Application for testing Compass Runtime Agent")
	gs.Require().NoError(err)

	synchronizedCompassAppName := fmt.Sprintf("mp-%s", compassAppName)

	applicationInterface := gs.applicationsClientSet.ApplicationconnectorV1alpha1().Applications()
	err = gs.assignApplicationToFormationAndWaitForSync(applicationInterface, synchronizedCompassAppName, applicationID)
	gs.Assert().NoError(err)

	// Compare Application created by Compass Runtime Agent with expected result
	err = gs.appComparator.Compare(expectedAppName, synchronizedCompassAppName)
	gs.Assert().NoError(err)

	// Clean up
	err = gs.directorClient.UnassignApplication(applicationID, gs.formationName)
	gs.Assert().NoError(err)

	err = gs.directorClient.UnregisterApplication(applicationID)
	gs.Require().NoError(err)
}

func (gs *CompassRuntimeAgentSuite) assignApplicationToFormationAndWaitForSync(appReader ApplicationReader, compassAppName, applicationID string) error {

	exec := func() error {
		return gs.directorClient.AssignApplicationToFormation(applicationID, gs.formationName)
	}

	verify := func() bool {
		_, err := appReader.Get(context.Background(), compassAppName, v1.GetOptions{})
		if err != nil {
			gs.T().Log(fmt.Sprintf("Failed to get app: %v", err))
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
