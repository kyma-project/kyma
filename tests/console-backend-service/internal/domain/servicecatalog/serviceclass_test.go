// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/wait"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/resource"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	rafterCli, _, err := client.NewDynamicClientWithConfig()
	require.NoError(t, err)

	// t.Log("Setup test service")
	// host, err := mockice.Start(rafterCli, TestNamespace, MockiceSvcName)
	// require.NoError(t, err)
	// defer mockice.Stop(rafterCli, TestNamespace, MockiceSvcName)

	assetGroupClient := resource.NewAssetGroup(rafterCli, expectedResource.Namespace, t.Logf)

	t.Log(fmt.Sprintf("Create AssetGroup %s", expectedResource.Name))
	err = assetGroupClient.Create(fixAssetGroupMeta(expectedResource.Name), fixCommonAssetGroupSpec())
	require.NoError(t, err)

	t.Log(fmt.Sprintf("Wait for AssetGroup %s Ready", expectedResource.Name))
	err = wait.ForAssetGroupReady(expectedResource.Name, assetGroupClient.Get)
	require.NoError(t, err)

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
		assetGroup {
			name
			namespace
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

	t.Log(fmt.Sprintf("Delete AssetGroup %s", expectedResource.Name))
	err = assetGroupClient.Delete(expectedResource.Name)
	require.NoError(t, err)

	t.Log(fmt.Sprintf("Wait for AssetGroup %s Deletion", expectedResource.Name))
	err = wait.ForAssetGroupDeletion(expectedResource.Name, assetGroupClient.Get)
	require.NoError(t, err)

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

	// AssetGroup
	require.NotEmpty(t, actual.AssetGroup)
	checkAssetGroup(t, fixture.AssetGroup(expected.Namespace, expected.Name), actual.AssetGroup)
}

func checkPlan(t *testing.T, expected, actual shared.ServicePlan) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// ExternalName
	assert.Equal(t, expected.ExternalName, actual.ExternalName)

	// RelatedServiceClassName
	assert.Equal(t, expected.RelatedServiceClassName, actual.RelatedServiceClassName)
}

func checkAssetGroup(t *testing.T, expected, actual shared.AssetGroup) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// Namespace
	assert.Equal(t, expected.Namespace, actual.Namespace)

	// DisplayName
	assert.Equal(t, expected.DisplayName, actual.DisplayName)

	// Description
	assert.Equal(t, expected.Description, actual.Description)
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

func fixAssetGroupMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
	}
}

func fixCommonAssetGroupSpec() v1beta1.CommonAssetGroupSpec {
	return v1beta1.CommonAssetGroupSpec{
		DisplayName: "Asset Group Sample",
		Description: "Asset Group Description",
		Sources: []v1beta1.Source{
			{
				Type: "markdown",
				Name: "markdown",
				Mode: v1beta1.AssetGroupSingle,
				URL:  "https://raw.githubusercontent.com/kyma-project/kyma/master/README.md",
			},
		},
	}
}
