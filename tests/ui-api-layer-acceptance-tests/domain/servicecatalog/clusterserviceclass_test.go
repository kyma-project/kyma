// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ClusterServicePlan struct {
	Name                           string
	DisplayName                    string
	ExternalName                   string
	Description                    string
	RelatedClusterServiceClassName string
	InstanceCreateParameterSchema  map[string]interface{}
}

type ClusterServiceClass struct {
	Name                string
	ExternalName        string
	DisplayName         string
	CreationTimestamp   int
	Description         string
	LongDescription     string
	ImageUrl            string
	DocumentationUrl    string
	SupportUrl          string
	ProviderDisplayName string
	Tags                []string
	Activated           bool
	Plans               []ClusterServicePlan
	apiSpec             map[string]interface{}
	asyncApiSpec        map[string]interface{}
	content             map[string]interface{}
}

type classesQueryResponse struct {
	ClusterServiceClasses []ClusterServiceClass
}

type classQueryResponse struct {
	ClusterServiceClass ClusterServiceClass
}

func TestClusterServiceClassesQueries(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}

	c, err := graphql.New()
	require.NoError(t, err)

	expectedResource := class()
	resourceDetailsQuery := `
		name
		externalName
		displayName
		creationTimestamp
		description
		longDescription
		imageUrl
		documentationUrl
		supportUrl
		providerDisplayName
		tags
		plans {
			name
			displayName
			externalName
			description
			relatedClusterServiceClassName
			instanceCreateParameterSchema
		}
		activated
	`

	t.Run("MultipleResources", func(t *testing.T) {
		query := fmt.Sprintf(`
			query {
				clusterServiceClasses {
					%s
				}
			}	
		`, resourceDetailsQuery)

		var res classesQueryResponse
		err = c.DoQuery(query, &res)

		require.NoError(t, err)
		assertClassExistsAndEqual(t, res.ClusterServiceClasses, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		query := fmt.Sprintf(`
			query($name: String!) {
				clusterServiceClass(name: $name) {
					%s
				}
			}	
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("name", expectedResource.Name)

		var res classQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkClass(t, expectedResource, res.ClusterServiceClass)
	})

}

func checkClass(t *testing.T, expected, actual ClusterServiceClass) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.ExternalName, actual.ExternalName)

	// Plans
	require.NotEmpty(t, actual.Plans)
	assertPlanExistsAndEqual(t, actual.Plans, expected.Plans[0])
}

func checkPlan(t *testing.T, expected, actual ClusterServicePlan) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.ExternalName, actual.ExternalName)
	assert.Equal(t, expected.RelatedClusterServiceClassName, actual.RelatedClusterServiceClassName)
}

func assertClassExistsAndEqual(t *testing.T, arr []ClusterServiceClass, expectedElement ClusterServiceClass) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkClass(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func assertPlanExistsAndEqual(t *testing.T, arr []ClusterServicePlan, expectedElement ClusterServicePlan) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkPlan(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func class() ClusterServiceClass {
	return ClusterServiceClass{
		Name:         "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
		ExternalName: "user-provided-service",
		Activated:    false,
		Plans: []ClusterServicePlan{
			{
				Name:                           "86064792-7ea2-467b-af93-ac9694d96d52",
				ExternalName:                   "default",
				RelatedClusterServiceClassName: "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
			},
		},
	}
}
