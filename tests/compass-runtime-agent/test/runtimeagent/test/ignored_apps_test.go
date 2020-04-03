package test

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	compassManagedApplicationName = "compass-managed-app"
	standAloneApplicationName     = "compass-not-managed-app"
)

func TestCompassRuntimeAgentNotManagedApplications(t *testing.T) {

	t.Run("should not delete an Application if it has no CompassMetadata in Spec", func(t *testing.T) {
		// when
		compassManagedApplicationTemplate := createSimpleApplicationTemplate(compassManagedApplicationName)
		compassManagedApplicationTemplate = withCompassMetadata(compassManagedApplicationTemplate, "App1")

		t.Logf("Creating Application %s managed by Compass Runtime Agent", compassManagedApplicationName)
		compassManagedApplication, err := testSuite.ApplicationCRClient.Create(compassManagedApplicationTemplate)
		require.NoError(t, err)
		defer func() {
			// In case test failed before deleting app, perform cleanup
			_ = testSuite.ApplicationCRClient.Delete(compassManagedApplicationName, &metav1.DeleteOptions{})
		}()

		standAloneApplicationTemplate := createSimpleApplicationTemplate(standAloneApplicationName)
		t.Logf("Creating stand alone Application %s not managed by Compass Runtime Agent", standAloneApplicationName)
		compassNotManagedApplication, err := testSuite.ApplicationCRClient.Create(standAloneApplicationTemplate)
		require.NoError(t, err)
		defer func() {
			t.Logf("Deleting stand alone Application %s not managed by Compass Runtime Agent", standAloneApplicationName)
			err := testSuite.ApplicationCRClient.Delete(compassNotManagedApplication.Name, &metav1.DeleteOptions{})
			require.NoError(t, err)

			testSuite.K8sResourceChecker.AssertAppResourcesDeleted(t, compassNotManagedApplication.Name)
		}()

		waitForAgentToApplyConfig(t, testSuite)

		// then
		t.Logf("Asserting that Application managed by Compass Runtime Agent is deleted if does not exist in Director config")
		testSuite.K8sResourceChecker.AssertAppResourcesDeleted(t, compassManagedApplication.Name)

		t.Logf("Asserting that Application not managed by Compass Runtime Agent still exists even if does not exist in Director config")
		returnedCompassNotManagedApplication, err := testSuite.ApplicationCRClient.Get(compassNotManagedApplication.Name, metav1.GetOptions{})
		require.NoError(t, err)
		assert.NotNil(t, returnedCompassNotManagedApplication)
		assert.Empty(t, returnedCompassNotManagedApplication.Spec.Services)
		assert.Nil(t, returnedCompassNotManagedApplication.Spec.CompassMetadata)
	})
}

func createSimpleApplicationTemplate(id string) *v1alpha1.Application {
	return &v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
		Spec: v1alpha1.ApplicationSpec{
			Description: "Description",
			Services:    []v1alpha1.Service{},
		},
	}
}

func withCompassMetadata(app *v1alpha1.Application, ids ...string) *v1alpha1.Application {
	app.Spec.CompassMetadata = &v1alpha1.CompassMetadata{Authentication: v1alpha1.Authentication{ClientIds: ids}}
	return app
}
