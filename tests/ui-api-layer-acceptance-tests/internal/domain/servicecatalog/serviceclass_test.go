// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/domain/shared"

	tester "github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serviceClassesQueryResponse struct {
	ServiceClasses []shared.ServiceClass
}

type serviceClassQueryResponse struct {
	ServiceClass shared.ServiceClass
}

func TestServiceClassesQueries(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	expectedResource := serviceClass()
	resourceDetailsQuery := `
		name
		environment
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
			relatedServiceClassName
			instanceCreateParameterSchema
			bindingCreateParameterSchema
		}
	`

	t.Run("MultipleResources", func(t *testing.T) {
		query := fmt.Sprintf(`
			query ($environment: String!) {
				serviceClasses(environment: $environment) {
					%s
				}
			}	
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("environment", expectedResource.Environment)

		var res serviceClassesQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		assertClassExistsAndEqual(t, res.ServiceClasses, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		query := fmt.Sprintf(`
			query ($name: String!, $environment: String!) {
				serviceClass(name: $name, environment: $environment) {
					%s
				}
			}	
		`, resourceDetailsQuery)
		req := graphql.NewRequest(query)
		req.SetVar("name", expectedResource.Name)
		req.SetVar("environment", expectedResource.Environment)

		var res serviceClassQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkClass(t, expectedResource, res.ServiceClass)
	})
}

func checkClass(t *testing.T, expected, actual shared.ServiceClass) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// Environment
	assert.Equal(t, expected.Environment, actual.Environment)

	// ExternalName
	assert.Equal(t, expected.ExternalName, actual.ExternalName)

	// Plans
	require.NotEmpty(t, actual.Plans)
	assertPlanExistsAndEqual(t, actual.Plans, expected.Plans[0])
}

func checkPlan(t *testing.T, expected, actual shared.ServicePlan) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// ExternalName
	assert.Equal(t, expected.ExternalName, actual.ExternalName)

	// RelatedServiceClassName
	assert.Equal(t, expected.RelatedServiceClassName, actual.RelatedServiceClassName)
}

func assertClassExistsAndEqual(t *testing.T, arr []shared.ServiceClass, expectedElement shared.ServiceClass) {
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

func assertPlanExistsAndEqual(t *testing.T, arr []shared.ServicePlan, expectedElement shared.ServicePlan) {
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

func serviceClass() shared.ServiceClass {
	return shared.ServiceClass{
		Name:         "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
		Environment:  tester.DefaultNamespace,
		ExternalName: "user-provided-service",
		Activated:    false,
		Plans: []shared.ServicePlan{
			{
				Name:                    "86064792-7ea2-467b-af93-ac9694d96d52",
				ExternalName:            "default",
				RelatedServiceClassName: "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
			},
		},
	}
}
