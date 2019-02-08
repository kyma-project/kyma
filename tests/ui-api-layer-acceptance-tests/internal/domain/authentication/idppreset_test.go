// +build acceptance

package authentication

import (
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/dex"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/module"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/client"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDPPresetQueriesAndMutations(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

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
}
