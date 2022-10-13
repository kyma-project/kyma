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

func (cs *CompassRuntimeAgentSuite) TestCreatingApplications() {
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
	err = cs.appComparator.Compare(expectedAppName, synchronizedCompassAppName)
	cs.Assert().NoError(err)

	// Clean up
	err = cs.directorClient.UnassignApplication(applicationID, cs.formationName)
	cs.Assert().NoError(err)

	err = cs.directorClient.UnregisterApplication(applicationID)
	cs.Require().NoError(err)
}

func (cs *CompassRuntimeAgentSuite) assignApplicationToFormationAndWaitForSync(appReader ApplicationReader, compassAppName, applicationID string) error {

	exec := func() error {
		return cs.directorClient.AssignApplicationToFormation(applicationID, cs.formationName)
	}

	verify := func() bool {
		_, err := appReader.Get(context.Background(), compassAppName, v1.GetOptions{})
		if err != nil {
			cs.T().Log(fmt.Sprintf("Failed to get app: %v", err))
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
