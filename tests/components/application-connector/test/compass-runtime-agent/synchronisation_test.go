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

// TODO: those values needs to be carefully picked to be in line with Compass Runtime Agent's configuration
const checkAppExistsPeriod = 30 * time.Second
const appCreationTimeout = 2 * time.Minute

type ApplicationReader interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Application, error)
}

func (ts *CompassRuntimeAgentSuite) TestCreatingApplications() {

	// Created in chart
	expectedAppName := "app1"
	compassAppName := expectedAppName + random.RandomString(10)

	// Create Application in Director and wait until it gets created
	applicationInterface := ts.applicationsClientSet.ApplicationconnectorV1alpha1().Applications()
	runtimeID, err := ts.createAppAndWaitForSync(applicationInterface, compassAppName, expectedAppName)
	ts.Require().NoError(err)

	// Compare Application created by Compass Runtime Agent with expected result
	err = ts.appComparator.Compare(expectedAppName, compassAppName)
	ts.Require().NoError(err)

	// Clean up
	err = ts.directorClient.UnregisterApplication(runtimeID)
	ts.Require().NoError(err)
}

func (ts *CompassRuntimeAgentSuite) createAppAndWaitForSync(appReader ApplicationReader, compassAppName, expectedAppName string) (string, error) {

	var runtimeID string

	exec := func() error {
		id, err := ts.directorClient.RegisterApplication(compassAppName)
		if err != nil {
			runtimeID = id
		}
		return err
	}

	verify := func() bool {
		_, err := appReader.Get(context.Background(), expectedAppName, v1.GetOptions{})
		if err != nil {
			ts.T().Log(fmt.Sprintf("Failed to get app: %v", err))
		}

		return err != nil
	}

	return runtimeID, executor.ExecuteAndWaitForCondition{
		RetryableExecuteFunc: exec,
		ConditionMetFunc:     verify,
		Tick:                 checkAppExistsPeriod,
		Timeout:              appCreationTimeout,
	}.Do()
}
