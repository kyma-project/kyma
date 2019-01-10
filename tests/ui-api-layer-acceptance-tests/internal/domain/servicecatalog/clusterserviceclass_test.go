// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/domain/shared"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type clusterServiceClassesQueryResponse struct {
	ClusterServiceClasses []shared.ClusterServiceClass
}

type clusterServiceClassQueryResponse struct {
	ClusterServiceClass shared.ClusterServiceClass
}

func TestClusterServiceClassesQueries(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	expectedResource := clusterServiceClass()
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
		labels
		activated
		plans {
			name
			displayName
			externalName
			description
			relatedClusterServiceClassName
			instanceCreateParameterSchema
		}
	`

	t.Run("MultipleResources", func(t *testing.T) {
		query := fmt.Sprintf(`
			query {
				clusterServiceClasses {
					%s
				}
			}	
		`, resourceDetailsQuery)

		var res clusterServiceClassesQueryResponse
		err = c.DoQuery(query, &res)

		require.NoError(t, err)
		assertClusterClassExistsAndEqual(t, res.ClusterServiceClasses, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		query := fmt.Sprintf(`
			query ($name: String!) {
				clusterServiceClass(name: $name) {
					%s
				}
			}	
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("name", expectedResource.Name)

		var res clusterServiceClassQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkClusterClass(t, expectedResource, res.ClusterServiceClass)
	})
}

func checkClusterClass(t *testing.T, expected, actual shared.ClusterServiceClass) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// ExternalName
	assert.Equal(t, expected.ExternalName, actual.ExternalName)

	// Plans
	require.NotEmpty(t, actual.Plans)
	assertClusterPlanExistsAndEqual(t, actual.Plans, expected.Plans[0])
}

func checkClusterPlan(t *testing.T, expected, actual shared.ClusterServicePlan) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// ExternalName
	assert.Equal(t, expected.ExternalName, actual.ExternalName)

	// RelatedClusterServiceClassName
	assert.Equal(t, expected.RelatedClusterServiceClassName, actual.RelatedClusterServiceClassName)
}

func assertClusterClassExistsAndEqual(t *testing.T, arr []shared.ClusterServiceClass, expectedElement shared.ClusterServiceClass) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkClusterClass(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func assertClusterPlanExistsAndEqual(t *testing.T, arr []shared.ClusterServicePlan, expectedElement shared.ClusterServicePlan) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkClusterPlan(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func clusterServiceClass() shared.ClusterServiceClass {
	return shared.ClusterServiceClass{
		Name:         "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
		ExternalName: "user-provided-service",
		Activated:    false,
		Plans: []shared.ClusterServicePlan{
			{
				Name:                           "86064792-7ea2-467b-af93-ac9694d96d52",
				ExternalName:                   "default",
				RelatedClusterServiceClassName: "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
			},
		},
	}
}
