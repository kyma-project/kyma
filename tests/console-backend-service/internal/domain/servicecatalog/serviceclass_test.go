// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"

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
		namespace
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
			relatedServiceClassName
			instanceCreateParameterSchema
			bindingCreateParameterSchema
		}
		apiSpec
		openApiSpec
		odataSpec
		asyncApiSpec
		content
	`

	t.Run("MultipleResources", func(t *testing.T) {
		req := fixServiceClassesRequest(resourceDetailsQuery, expectedResource)

		var res serviceClassesQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		assertClassExistsAndEqual(t, res.ServiceClasses, expectedResource)
	})

	t.Run("SingleResource", func(t *testing.T) {
		req := fixServiceClassRequest(resourceDetailsQuery, expectedResource)

		var res serviceClassQueryResponse
		err = c.Do(req, &res)

		require.NoError(t, err)
		checkClass(t, expectedResource, res.ServiceClass)
	})

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:  {fixServiceClassRequest(resourceDetailsQuery, expectedResource)},
		auth.List: {fixServiceClassesRequest(resourceDetailsQuery, expectedResource)},
	}
	AuthSuite.Run(t, ops)
}

func checkClass(t *testing.T, expected, actual shared.ServiceClass) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// Namespace
	assert.Equal(t, expected.Namespace, actual.Namespace)

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
	className := fixture.TestingBundleClassName
	return shared.ServiceClass{
		Name:         className,
		Namespace:    TestNamespace,
		ExternalName: fixture.TestingBundleClassExternalName,
		Activated:    false,
		Plans: []shared.ServicePlan{
			{
				Name:                    fixture.TestingBundleMinimalPlanName,
				ExternalName:            fixture.TestingBundleMinimalPlanExternalName,
				RelatedServiceClassName: className,
			},
			{
				Name:                    fixture.TestingBundleFullPlanName,
				ExternalName:            fixture.TestingBundleFullPlanExternalName,
				RelatedServiceClassName: className,
			},
		},
	}
}

func fixServiceClassRequest(resourceDetailsQuery string, expectedResource shared.ServiceClass) *graphql.Request {
	query := fmt.Sprintf(`
			query ($name: String!, $namespace: String!) {
				serviceClass(name: $name, namespace: $namespace) {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)
	req.SetVar("namespace", expectedResource.Namespace)

	return req
}

func fixServiceClassesRequest(resourceDetailsQuery string, expectedResource shared.ServiceClass) *graphql.Request {
	query := fmt.Sprintf(`
			query ($namespace: String!) {
				serviceClasses(namespace: $namespace) {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", expectedResource.Namespace)

	return req
}
