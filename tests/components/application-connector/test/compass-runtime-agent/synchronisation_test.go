package compass_runtime_agent

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const checkAppExistsPeriod = 30 * time.Second
const appCreationTimeout = 2 * time.Minute

type DirectorClient interface {
	CreateApplication(name string) (string, error)
	DeleteApplication(id string) error
}

type AppComparator interface {
	Compare(assertions *require.Assertions, actualApp, expectedApp string) error
}

type ApplicationReader interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Application, error)
}

func NewDirectorClient() DirectorClient {
	return nil
}

func (gs *CompassRuntimeAgentSuite) TestCreatingApplications() {
	// Created in chart
	expectedAppName := "app1"
	compassAppName := expectedAppName + RandomString(10)
	directorClient := NewDirectorClient()

	// Create Application in Director and wait until it gets created
	applicationInterface := gs.cli.ApplicationconnectorV1alpha1().Applications()
	runtimeID, err := gs.createAppAndWaitForSync(directorClient, applicationInterface, compassAppName, expectedAppName)
	gs.Require().NoError(err)

	// Compare Application created by Compass Runtime Agent with expected result
	err = gs.appComparator.Compare(gs.Require(), compassAppName, expectedAppName)
	gs.Require().NoError(err)

	// Clean up
	err = directorClient.DeleteApplication(runtimeID)
	gs.Require().NoError(err)
}

func (gs *CompassRuntimeAgentSuite) createAppAndWaitForSync(directorClient DirectorClient, appReader ApplicationReader, compassAppName, expectedAppName string) (string, error) {

	var runtimeID string

	exec := func() error {
		id, err := directorClient.CreateApplication(compassAppName)
		if err != nil {
			runtimeID = id
		}
		return err
	}

	verify := func() bool {
		_, err := appReader.Get(context.Background(), expectedAppName, v1.GetOptions{})
		if err != nil {
			gs.T().Log(fmt.Sprintf("Failed to get app: %v", err))
		}

		return err != nil
	}

	return runtimeID, ExecuteAndWaitForCondition{
		exec,
		verify,
		checkAppExistsPeriod,
		appCreationTimeout,
	}.Do()
}
