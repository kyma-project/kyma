package k8s_test

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/automock"
	scMock "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/shared/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/api/apps/v1beta2"
)

func TestDeploymentResolver_DeploymentsQuery(t *testing.T) {
	environment := "test"

	t.Run("Success with default", func(t *testing.T) {
		deployment := fixDeployment("test", environment, "function")
		deployments := []*v1beta2.Deployment{deployment, deployment}

		expected := gqlschema.Deployment{
			Name:        "test",
			Environment: environment,
			Labels: gqlschema.Labels{
				"function": "",
			},
		}

		svc := automock.NewDeploymentLister()
		svc.On("List", environment).Return(deployments, nil).Once()
		svc.On("ListWithoutFunctions", mock.Anything, mock.Anything).Return(deployments, nil).Once()
		resolver := k8s.NewDeploymentResolver(svc, nil, nil)

		result, err := resolver.DeploymentsQuery(nil, environment, nil)

		require.NoError(t, err)
		assert.Equal(t, []gqlschema.Deployment{expected, expected}, result)
		svc.AssertNotCalled(t, "ListWithoutFunctions", mock.Anything, mock.Anything)
	})

	t.Run("Success with functions", func(t *testing.T) {
		deployment := fixDeployment("test", environment, "deployment")
		deployments := []*v1beta2.Deployment{deployment, deployment}

		expected := gqlschema.Deployment{
			Name:        "test",
			Environment: environment,
			Labels: gqlschema.Labels{
				"deployment": "",
			},
		}

		svc := automock.NewDeploymentLister()
		svc.On("List", environment).Return(deployments, nil).Once()
		svc.On("ListWithoutFunctions", mock.Anything, mock.Anything).Return(deployments, nil).Once()
		resolver := k8s.NewDeploymentResolver(svc, nil, nil)

		result, err := resolver.DeploymentsQuery(nil, environment, getBoolPointer(false))

		require.NoError(t, err)
		assert.Equal(t, []gqlschema.Deployment{expected, expected}, result)
		svc.AssertNotCalled(t, "ListWithoutFunctions", mock.Anything, mock.Anything)
	})

	t.Run("Success without functions", func(t *testing.T) {
		deployment := fixDeployment("test", environment, "function")
		deployments := []*v1beta2.Deployment{deployment, deployment}

		expected := gqlschema.Deployment{
			Name:        "test",
			Environment: environment,
			Labels: gqlschema.Labels{
				"function": "",
			},
		}

		svc := automock.NewDeploymentLister()
		svc.On("List", mock.Anything, mock.Anything).Return(deployments, nil).Once()
		svc.On("ListWithoutFunctions", environment).Return(deployments, nil).Once()
		resolver := k8s.NewDeploymentResolver(svc, nil, nil)

		result, err := resolver.DeploymentsQuery(nil, environment, getBoolPointer(true))

		require.NoError(t, err)
		assert.Equal(t, []gqlschema.Deployment{expected, expected}, result)
		svc.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
	})

	t.Run("Not found with functions", func(t *testing.T) {
		svc := automock.NewDeploymentLister()
		svc.On("List", environment).Return([]*v1beta2.Deployment{}, nil).Once()
		svc.On("ListWithoutFunctions", mock.Anything, mock.Anything).Return([]*v1beta2.Deployment{}, nil).Once()
		resolver := k8s.NewDeploymentResolver(svc, nil, nil)

		result, err := resolver.DeploymentsQuery(nil, environment, getBoolPointer(false))

		require.NoError(t, err)
		assert.Empty(t, result)
		svc.AssertNotCalled(t, "ListWithoutFunctions", mock.Anything, mock.Anything)
	})

	t.Run("Not found without functions", func(t *testing.T) {
		svc := automock.NewDeploymentLister()
		svc.On("List", mock.Anything, mock.Anything).Return([]*v1beta2.Deployment{}, nil).Once()
		svc.On("ListWithoutFunctions", environment).Return([]*v1beta2.Deployment{}, nil).Once()
		resolver := k8s.NewDeploymentResolver(svc, nil, nil)

		result, err := resolver.DeploymentsQuery(nil, environment, getBoolPointer(true))

		require.NoError(t, err)
		assert.Empty(t, result)
		svc.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
	})

	t.Run("Error with functions", func(t *testing.T) {
		svc := automock.NewDeploymentLister()
		svc.On("List", environment).Return(nil, errors.New("test")).Once()
		defer svc.AssertExpectations(t)
		resolver := k8s.NewDeploymentResolver(svc, nil, nil)

		_, err := resolver.DeploymentsQuery(nil, environment, getBoolPointer(false))

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Error without functions", func(t *testing.T) {
		svc := automock.NewDeploymentLister()
		svc.On("ListWithoutFunctions", environment).Return(nil, errors.New("test")).Once()
		defer svc.AssertExpectations(t)
		resolver := k8s.NewDeploymentResolver(svc, nil, nil)

		_, err := resolver.DeploymentsQuery(nil, environment, getBoolPointer(true))

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestDeploymentResolver_DeploymentBoundServiceInstanceNamesField(t *testing.T) {
	environment := "test"

	t.Run("Success for deployment", func(t *testing.T) {
		deployment := &gqlschema.Deployment{
			Name:        "test",
			Environment: environment,
			Labels:      gqlschema.Labels{},
		}

		usage := &v1alpha1.ServiceBindingUsage{
			Spec: v1alpha1.ServiceBindingUsageSpec{
				ServiceBindingRef: v1alpha1.LocalReferenceByName{
					Name: "test",
				},
			},
		}

		binding := &v1beta1.ServiceBinding{
			Spec: v1beta1.ServiceBindingSpec{
				ServiceInstanceRef: v1beta1.LocalObjectReference{
					Name: "instance",
				},
			},
		}

		lister := new(scMock.ServiceBindingUsageLister)
		lister.On("ListForDeployment", deployment.Environment, "deployment", deployment.Name).Return([]*v1alpha1.ServiceBindingUsage{usage}, nil)
		getter := new(scMock.ServiceBindingFinderLister)
		getter.On("Find", deployment.Environment, usage.Spec.ServiceBindingRef.Name).Return(binding, nil)

		scRetriever := new(scMock.ServiceCatalogRetriever)
		scRetriever.On("ServiceBinding").Return(getter)

		scaRetriever := new(scMock.ServiceCatalogAddonsRetriever)
		scaRetriever.On("ServiceBindingUsage").Return(lister)

		resolver := k8s.NewDeploymentResolver(nil, scRetriever, scaRetriever)

		result, err := resolver.DeploymentBoundServiceInstanceNamesField(nil, deployment)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"instance",
		}, result)
	})

	t.Run("Success for function", func(t *testing.T) {
		deployment := &gqlschema.Deployment{
			Name:        "test",
			Environment: environment,
			Labels: gqlschema.Labels{
				"function": "",
			},
		}

		usage := &v1alpha1.ServiceBindingUsage{
			Spec: v1alpha1.ServiceBindingUsageSpec{
				ServiceBindingRef: v1alpha1.LocalReferenceByName{
					Name: "test",
				},
			},
		}

		binding := &v1beta1.ServiceBinding{
			Spec: v1beta1.ServiceBindingSpec{
				ServiceInstanceRef: v1beta1.LocalObjectReference{
					Name: "instance",
				},
			},
		}

		lister := new(scMock.ServiceBindingUsageLister)
		lister.On("ListForDeployment", deployment.Environment, "function", deployment.Name).Return([]*v1alpha1.ServiceBindingUsage{usage}, nil)
		getter := new(scMock.ServiceBindingFinderLister)
		getter.On("Find", deployment.Environment, usage.Spec.ServiceBindingRef.Name).Return(binding, nil)

		scRetriever := new(scMock.ServiceCatalogRetriever)
		scRetriever.On("ServiceBinding").Return(getter)

		scaRetriever := new(scMock.ServiceCatalogAddonsRetriever)
		scaRetriever.On("ServiceBindingUsage").Return(lister)

		resolver := k8s.NewDeploymentResolver(nil, scRetriever, scaRetriever)

		result, err := resolver.DeploymentBoundServiceInstanceNamesField(nil, deployment)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"instance",
		}, result)
	})

	t.Run("No usages", func(t *testing.T) {
		deployment := &gqlschema.Deployment{
			Name:        "test",
			Environment: environment,
			Labels: gqlschema.Labels{
				"function": "",
			},
		}

		lister := new(scMock.ServiceBindingUsageLister)
		lister.On("ListForDeployment", deployment.Environment, "function", deployment.Name).Return([]*v1alpha1.ServiceBindingUsage{}, nil)

		scaRetriever := new(scMock.ServiceCatalogAddonsRetriever)
		scaRetriever.On("ServiceBindingUsage").Return(lister)

		resolver := k8s.NewDeploymentResolver(nil, nil, scaRetriever)

		result, err := resolver.DeploymentBoundServiceInstanceNamesField(nil, deployment)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("No binding", func(t *testing.T) {
		deployment := &gqlschema.Deployment{
			Name:        "test",
			Environment: environment,
			Labels: gqlschema.Labels{
				"function": "",
			},
		}

		usage := &v1alpha1.ServiceBindingUsage{
			Spec: v1alpha1.ServiceBindingUsageSpec{
				ServiceBindingRef: v1alpha1.LocalReferenceByName{
					Name: "test",
				},
			},
		}

		lister := new(scMock.ServiceBindingUsageLister)
		lister.On("ListForDeployment", deployment.Environment, "function", deployment.Name).Return([]*v1alpha1.ServiceBindingUsage{usage}, nil)
		getter := new(scMock.ServiceBindingFinderLister)
		getter.On("Find", deployment.Environment, usage.Spec.ServiceBindingRef.Name).Return(nil, nil)

		scRetriever := new(scMock.ServiceCatalogRetriever)
		scRetriever.On("ServiceBinding").Return(getter)

		scaRetriever := new(scMock.ServiceCatalogAddonsRetriever)
		scaRetriever.On("ServiceBindingUsage").Return(lister)

		resolver := k8s.NewDeploymentResolver(nil, scRetriever, scaRetriever)

		result, err := resolver.DeploymentBoundServiceInstanceNamesField(nil, deployment)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error when deployment not provided", func(t *testing.T) {
		resolver := k8s.NewDeploymentResolver(nil, nil, nil)

		_, err := resolver.DeploymentBoundServiceInstanceNamesField(nil, nil)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Error while listing usages", func(t *testing.T) {
		deployment := &gqlschema.Deployment{
			Name:        "test",
			Environment: environment,
			Labels: gqlschema.Labels{
				"function": "",
			},
		}

		lister := new(scMock.ServiceBindingUsageLister)
		lister.On("ListForDeployment", deployment.Environment, "function", deployment.Name).Return([]*v1alpha1.ServiceBindingUsage{}, errors.New("trolololo"))
		defer lister.AssertExpectations(t)

		scaRetriever := new(scMock.ServiceCatalogAddonsRetriever)
		scaRetriever.On("ServiceBindingUsage").Return(lister)

		resolver := k8s.NewDeploymentResolver(nil, nil, scaRetriever)

		_, err := resolver.DeploymentBoundServiceInstanceNamesField(nil, deployment)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Error while getting binding", func(t *testing.T) {
		deployment := &gqlschema.Deployment{
			Name:        "test",
			Environment: environment,
			Labels: gqlschema.Labels{
				"function": "",
			},
		}

		usage := &v1alpha1.ServiceBindingUsage{
			Spec: v1alpha1.ServiceBindingUsageSpec{
				ServiceBindingRef: v1alpha1.LocalReferenceByName{
					Name: "test",
				},
			},
		}

		lister := new(scMock.ServiceBindingUsageLister)
		lister.On("ListForDeployment", deployment.Environment, "function", deployment.Name).Return([]*v1alpha1.ServiceBindingUsage{usage}, nil)
		defer lister.AssertExpectations(t)
		getter := new(scMock.ServiceBindingFinderLister)
		getter.On("Find", deployment.Environment, usage.Spec.ServiceBindingRef.Name).Return(nil, errors.New("trolololo"))
		defer getter.AssertExpectations(t)

		scRetriever := new(scMock.ServiceCatalogRetriever)
		scRetriever.On("ServiceBinding").Return(getter)

		scaRetriever := new(scMock.ServiceCatalogAddonsRetriever)
		scaRetriever.On("ServiceBindingUsage").Return(lister)

		resolver := k8s.NewDeploymentResolver(nil, scRetriever, scaRetriever)

		_, err := resolver.DeploymentBoundServiceInstanceNamesField(nil, deployment)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func getBoolPointer(value bool) *bool {
	return &value
}
