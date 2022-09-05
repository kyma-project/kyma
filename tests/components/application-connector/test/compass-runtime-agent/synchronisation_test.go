package compass_runtime_agent

import (
	"context"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type DirectorClient interface {
	CreateApplication(name string) error
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

	applicationInterface := gs.cli.ApplicationconnectorV1alpha1().Applications()
	err := createAppAndWaitForSync(directorClient, applicationInterface, compassAppName, expectedAppName)
	gs.Require().NoError(err)

	err = gs.appComparator.Compare(gs.Require(), compassAppName, expectedAppName)
	gs.Require().NoError(err)
}

func createAppAndWaitForSync(directorClient DirectorClient, appReader ApplicationReader, compassAppName, expectedAppName string) error {
	exec := func() error {
		return directorClient.CreateApplication(compassAppName)
	}

	verify := func() error {
		_, err := appReader.Get(context.Background(), expectedAppName, v1.GetOptions{})
		return err
	}

	return ExecuteAndWaitForCondition{
		exec,
		verify,
		10 * time.Second,
		1 * time.Minute,
	}.Do()
}
