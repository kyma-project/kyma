package application_mapping_controller

import (
	"context"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/application-mapping-controller/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

const (
	amName    = "testMapping"
	namespace = "testNamespace"
)

func TestAppMappingReconciler_Reconcile(t *testing.T) {
	t.Run("should deploy Gateway when first Application Mapping is created in namespace", func(t *testing.T) {
		//given
		gatewayDeployer := &mocks.GatewayDeployer{}
		gatewayDeployer.On("GatewayExists", namespace).Return(false)
		gatewayDeployer.On("DeployGateway", namespace).Return(nil)

		amClient := &mocks.ApplicationMappingManagerClient{}

		namespacedName := types.NamespacedName{
			Name:      amName,
			Namespace: namespace,
		}

		amClient.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.ApplicationMapping")).
			Run(appMap).Return(nil)

		amClient.On("List", context.Background(), mock.AnythingOfType("*v1alpha1.ApplicationMappingList"), mock.Anything).
			Run(appMapList).Return(nil)

		reconciler := NewReconciler(amClient, gatewayDeployer)

		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      amName,
			}}

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		gatewayDeployer.AssertExpectations(t)
	})

	t.Run("should remove Gateway when last Application Mapping is deleted from namespace", func(t *testing.T) {
		//given
		gatewayDeployer := &mocks.GatewayDeployer{}
		gatewayDeployer.On("RemoveGateway", namespace).Return(nil)

		amClient := &mocks.ApplicationMappingManagerClient{}

		namespacedName := types.NamespacedName{
			Name:      amName,
			Namespace: namespace,
		}

		amClient.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.ApplicationMapping")).Return(errors.NewNotFound(schema.GroupResource{}, amName))

		amClient.On("List", context.Background(), mock.AnythingOfType("*v1alpha1.ApplicationMappingList"), mock.Anything).Return(nil)

		reconciler := NewReconciler(amClient, gatewayDeployer)

		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      amName,
			}}

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		gatewayDeployer.AssertExpectations(t)
	})

	t.Run("should not deploy Gateway when second Application Mapping is created in namespace and Gateway already exists", func(t *testing.T) {
		//given
		gatewayDeployer := &mocks.GatewayDeployer{}
		gatewayDeployer.On("GatewayExists", namespace).Return(true)

		amClient := &mocks.ApplicationMappingManagerClient{}

		namespacedName := types.NamespacedName{
			Name:      amName,
			Namespace: namespace,
		}

		amClient.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.ApplicationMapping")).
			Run(appMap).Return(nil)

		amClient.On("List", context.Background(), mock.AnythingOfType("*v1alpha1.ApplicationMappingList"), mock.Anything).
			Run(appMapListWithTwoAppMappings).Return(nil)

		reconciler := NewReconciler(amClient, gatewayDeployer)

		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      amName,
			}}

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		gatewayDeployer.AssertExpectations(t)
	})
}

func appMapList(args mock.Arguments) {
	list := getAppMapList(args)
	list.Items = []v1alpha1.ApplicationMapping{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      amName,
				Namespace: namespace,
			},
		},
	}
}

func appMapListWithTwoAppMappings(args mock.Arguments) {
	list := getAppMapList(args)
	list.Items = []v1alpha1.ApplicationMapping{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      amName,
				Namespace: namespace,
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      amName,
				Namespace: namespace,
			},
		},
	}
}

func appMap(args mock.Arguments) {
	appMapInstance := getAppMap(args)
	appMapInstance.Name = amName
	appMapInstance.Namespace = namespace
}

func getAppMap(args mock.Arguments) *v1alpha1.ApplicationMapping {
	appMapInstance := args.Get(2).(*v1alpha1.ApplicationMapping)
	return appMapInstance
}

func getAppMapList(args mock.Arguments) *v1alpha1.ApplicationMappingList {
	appMapListInstance := args.Get(1).(*v1alpha1.ApplicationMappingList)
	return appMapListInstance
}
