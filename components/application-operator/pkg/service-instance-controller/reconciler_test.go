package service_instance_controller

import (
	"context"
	"testing"

	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	mocks2 "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/gateway/mocks"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/application-operator/pkg/service-instance-controller/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	svcInstanceName = "testInstance"
	namespace       = "testNamespace"
)

func TestServiceInstanceReconciler_Reconcile(t *testing.T) {
	namespacedName := types.NamespacedName{
		Name:      svcInstanceName,
		Namespace: namespace,
	}

	logger := logrus.WithField("controller", "Service Instance Tests")

	t.Run("should deploy Gateway when first Service Instance is created in namespace", func(t *testing.T) {
		//given
		gatewayDeployer := &mocks2.GatewayManager{}
		gatewayDeployer.On("GatewayExists", namespace).Return(false, release.Status_UNKNOWN, nil)
		gatewayDeployer.On("InstallGateway", namespace).Return(nil)

		amClient := &mocks.ServiceInstanceManagerClient{}

		amClient.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1beta1.ServiceInstance")).
			Run(serviceInstance).Return(nil)

		amClient.On("List", context.Background(), mock.AnythingOfType("*v1beta1.ServiceInstanceList"), &client.ListOptions{Namespace: namespace}).
			Run(serviceInstancesList).Return(nil)

		reconciler := NewReconciler(amClient, gatewayDeployer, logger)

		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      svcInstanceName,
			}}

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		gatewayDeployer.AssertExpectations(t)
	})

	t.Run("should remove Gateway when last Service Instance is deleted from namespace", func(t *testing.T) {
		//given
		gatewayDeployer := &mocks2.GatewayManager{}
		gatewayDeployer.On("DeleteGateway", namespace).Return(nil)

		amClient := &mocks.ServiceInstanceManagerClient{}

		amClient.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1beta1.ServiceInstance")).Return(errors.NewNotFound(schema.GroupResource{}, svcInstanceName))

		amClient.On("List", context.Background(), mock.AnythingOfType("*v1beta1.ServiceInstanceList"), &client.ListOptions{Namespace: namespace}).Return(nil)

		reconciler := NewReconciler(amClient, gatewayDeployer, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		gatewayDeployer.AssertExpectations(t)
	})

	t.Run("should not deploy Gateway when second Service Instance is created in namespace and Gateway already exists", func(t *testing.T) {
		//given
		gatewayDeployer := &mocks2.GatewayManager{}
		gatewayDeployer.On("GatewayExists", namespace).Return(true, release.Status_DEPLOYED, nil)

		amClient := &mocks.ServiceInstanceManagerClient{}

		amClient.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1beta1.ServiceInstance")).
			Run(serviceInstance).Return(nil)

		amClient.On("List", context.Background(), mock.AnythingOfType("*v1beta1.ServiceInstanceList"), &client.ListOptions{Namespace: namespace}).
			Run(listWithTwoServiceInstances).Return(nil)

		reconciler := NewReconciler(amClient, gatewayDeployer, logger)

		request := reconcile.Request{
			NamespacedName: namespacedName,
		}

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		gatewayDeployer.AssertExpectations(t)
	})

	t.Run("should not deploy Gateway in system namespaces", func(t *testing.T) {
		//given
		gatewayDeployer := &mocks2.GatewayManager{}

		amClient := &mocks.ServiceInstanceManagerClient{}

		reconciler := NewReconciler(amClient, gatewayDeployer, logger)

		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "kyma-system",
				Name:      svcInstanceName,
			}}

		//when
		_, err := reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		gatewayDeployer.AssertExpectations(t)
	})
}

func serviceInstancesList(args mock.Arguments) {
	list := getServiceInstanceList(args)
	list.Items = []v1beta1.ServiceInstance{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      svcInstanceName,
				Namespace: namespace,
			},
		},
	}
}

func listWithTwoServiceInstances(args mock.Arguments) {
	list := getServiceInstanceList(args)
	list.Items = []v1beta1.ServiceInstance{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      svcInstanceName,
				Namespace: namespace,
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      svcInstanceName,
				Namespace: namespace,
			},
		},
	}
}

func serviceInstance(args mock.Arguments) {
	serviceInstance := getServiceInstance(args)
	serviceInstance.Name = svcInstanceName
	serviceInstance.Namespace = namespace
}

func getServiceInstance(args mock.Arguments) *v1beta1.ServiceInstance {
	serviceInstance := args.Get(2).(*v1beta1.ServiceInstance)
	return serviceInstance
}

func getServiceInstanceList(args mock.Arguments) *v1beta1.ServiceInstanceList {
	serviceListInstance := args.Get(1).(*v1beta1.ServiceInstanceList)
	return serviceListInstance
}
