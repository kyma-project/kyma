// +build acceptance

package authentication

import (
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/module"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDPPresetQueriesAndMutations(t *testing.T) {
	t.Skip("skipping unstable test")

	c, err := graphql.New()
	require.NoError(t, err)

	module.SkipPluggableTestIfShould(t, c, ModuleName)

	client, _, err := client.NewIDPPresetClientWithConfig()
	require.NoError(t, err)

	expectedResource := idpPreset("test-name", "test-issuer", "https://test-jwksUri")
	resourceDetailsQuery := idpPresetDetailsFields()

	t.Log("Create IDP Preset")
	createRes, err := createIDPPreset(c, resourceDetailsQuery, expectedResource)

	require.NoError(t, err)
	checkIDPPreset(t, expectedResource, createRes.CreateIDPPreset)

	t.Log("Wait For IDP Preset Creation")
	err = waitForIDPPresetCreation(expectedResource.Name, client)
	assert.NoError(t, err)

	t.Log("Query Single Resource")
	res, err := querySingleIDPPreset(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	checkIDPPreset(t, expectedResource, res.IDPPreset)

	t.Log("Query Multiple Resources")
	multipleRes, err := queryMultipleIDPPresets(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	checkIDPPresets(t, expectedResource, multipleRes.IDPPresets)

	t.Log("Delete IDP Preset")
	deleteRes, err := deleteIDPPreset(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	checkIDPPreset(t, expectedResource, deleteRes.DeleteIDPPreset)

	t.Log("Wait For IDP Preset Deletion")
	err = waitForIDPPresetDeletion(expectedResource.Name, client)
	assert.NoError(t, err)

	t.Log("Checking authorization directives...")
	as := auth.New()
	ops := &auth.OperationsInput{
		auth.Get:    {singleResourceQueryRequest(resourceDetailsQuery, expectedResource)},
		auth.List:   {multipleResourcesQueryRequest(resourceDetailsQuery, expectedResource)},
		auth.Create: {fixCreateIDPPresetRequest(resourceDetailsQuery, idpPreset("", "", ""))},
		auth.Delete: {fixDeleteIDPPresetRequest(resourceDetailsQuery, expectedResource)},
	}
	as.Run(t, ops)
}
