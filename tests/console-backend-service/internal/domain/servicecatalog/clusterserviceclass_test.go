// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/wait"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type clusterServiceClassesQueryResponse struct {
	ClusterServiceClasses []shared.ClusterServiceClass
}

type clusterServiceClassQueryResponse struct {
	ClusterServiceClass shared.ClusterServiceClass
}

func TestClusterServiceClassesQueries(t *testing.T) {
	t.Skip("skipping unstable test")
	c, err := graphql.New()
	require.NoError(t, err)

	expectedResource := clusterServiceClass()

	cmsCli, _, err := client.NewDynamicClientWithConfig()
	require.NoError(t, err)

	clusterDocsTopicClient := resource.NewClusterDocsTopic(cmsCli, t.Logf)

	t.Log(fmt.Sprintf("Wait for clusterDocsTopic %s Ready", expectedResource.Name))
	err = wait.ForClusterDocsTopicReady(expectedResource.Name, clusterDocsTopicClient.Get)
	require.NoError(t, err)

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
		instances {
			name
			namespace
		}
		plans {
			name
			displayName
			externalName
			description
			relatedClusterServiceClassName
			instanceCreateParameterSchema
		}
		clusterDocsTopic {
			name
    		groupName
    		assets {
				name
				type
				files {
					url
					metadata
				}
			}
    		displayName
    		description
		}
	`

	t.Run("MultipleResources", func(t *testing.T) {
		req := fixClusterServiceClassesRequest(resourceDetailsQuery)

		var res clusterServiceClassesQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		assertClusterClassExistsAndEqual(t, res.ClusterServiceClasses, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		req := fixClusterServiceClassRequest(resourceDetailsQuery, expectedResource.Name)

		var res clusterServiceClassQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkClusterClass(t, expectedResource, res.ClusterServiceClass)
	})

	t.Log(fmt.Sprintf("Delete clusterDocsTopic %s", expectedResource.Name))
	err = clusterDocsTopicClient.Delete(expectedResource.Name)
	require.NoError(t, err)

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:  {fixClusterServiceClassRequest(resourceDetailsQuery, "test")},
		auth.List: {fixClusterServiceClassesRequest(resourceDetailsQuery)},
	}
	AuthSuite.Run(t, ops)
}

func checkClusterClass(t *testing.T, expected, actual shared.ClusterServiceClass) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// ExternalName
	assert.Equal(t, expected.ExternalName, actual.ExternalName)

	// Plans
	require.NotEmpty(t, actual.Plans)
	assertClusterPlanExistsAndEqual(t, actual.Plans, expected.Plans[0])

	// ClusterDocsTopic
	require.NotEmpty(t, actual.ClusterDocsTopic)
	checkClusterDocsTopic(t, fixTestingBundleClusterDocsTopic(), actual.ClusterDocsTopic)
}

func checkClusterPlan(t *testing.T, expected, actual shared.ClusterServicePlan) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// ExternalName
	assert.Equal(t, expected.ExternalName, actual.ExternalName)

	// RelatedClusterServiceClassName
	assert.Equal(t, expected.RelatedClusterServiceClassName, actual.RelatedClusterServiceClassName)
}

func checkClusterDocsTopic(t *testing.T, expected, actual shared.ClusterDocsTopic) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// DisplayName
	assert.Equal(t, expected.DisplayName, actual.DisplayName)

	// Description
	assert.Equal(t, expected.Description, actual.Description)
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
	className := fixture.TestingBundleClassName
	return shared.ClusterServiceClass{
		Name:         className,
		ExternalName: fixture.TestingBundleClassExternalName,
		Activated:    false,
		Plans: []shared.ClusterServicePlan{
			{
				Name:                           fixture.TestingBundleMinimalPlanName,
				ExternalName:                   fixture.TestingBundleMinimalPlanExternalName,
				RelatedClusterServiceClassName: className,
			},
			{
				Name:                           fixture.TestingBundleFullPlanName,
				ExternalName:                   fixture.TestingBundleFullPlanExternalName,
				RelatedClusterServiceClassName: className,
			},
		},
	}
}

func fixClusterServiceClassRequest(resourceDetailsQuery, name string) *graphql.Request {
	query := fmt.Sprintf(`
			query ($name: String!) {
				clusterServiceClass(name: $name) {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", name)

	return req
}

func fixClusterServiceClassesRequest(resourceDetailsQuery string) *graphql.Request {
	query := fmt.Sprintf(`
			query {
				clusterServiceClasses {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	return req
}

func fixClusterDocsTopicMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
	}
}

func fixTestingBundleClusterDocsTopic() shared.ClusterDocsTopic {
	return shared.ClusterDocsTopic{
		Name:        fixture.TestingBundleClassName,
		DisplayName: "Documentation for testing-0.0.1",
		Description: "Overall documentation",
	}
}
