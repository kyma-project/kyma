package accessservice

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"fmt"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/k8sconsts"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/accessservice/mocks"
)

var config = AccessServiceManagerConfig{
	TargetPort: 8081,
}

const (
	applicationUID = types.UID("appUID")
)

func TestAccessServiceManager_Create(t *testing.T) {

	t.Run("should create new access service", func(t *testing.T) {
		// given
		expectedService := mockService("ec-default", "uuid-1", "service-uuid1", 8081)
		serviceInterface := new(mocks.ServiceInterface)
		serviceInterface.On("Create", expectedService).Return(expectedService, nil)

		manager := NewAccessServiceManager(serviceInterface, config)

		// when
		err := manager.Create("ec-default", "appUID", "uuid-1", "service-uuid1")

		// then
		assert.NoError(t, err)
		serviceInterface.AssertExpectations(t)
	})

	t.Run("should handle errors", func(t *testing.T) {
		// given
		expectedService := mockService("ec-default", "uuid-1", "service-uuid1", 8081)
		serviceInterface := new(mocks.ServiceInterface)
		serviceInterface.On("Create", expectedService).Return(nil, errors.New("some error"))

		manager := NewAccessServiceManager(serviceInterface, config)

		// when
		err := manager.Create("ec-default", "appUID", "uuid-1", "service-uuid1")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceInterface.AssertExpectations(t)
	})
}

func TestAccessServiceManager_Upsert(t *testing.T) {

	t.Run("should create access service if not exists", func(t *testing.T) {
		// given
		expectedService := mockService("ec-default", "uuid-1", "service-uuid1", 8081)
		serviceInterface := new(mocks.ServiceInterface)
		serviceInterface.On("Create", expectedService).Return(expectedService, nil)

		manager := NewAccessServiceManager(serviceInterface, config)

		// when
		err := manager.Upsert("ec-default", "appUID", "uuid-1", "service-uuid1")

		// then
		assert.NoError(t, err)
		serviceInterface.AssertExpectations(t)
	})

	t.Run("should not fail if access service exists", func(t *testing.T) {
		// given
		expectedService := mockService("ec-default", "uuid-1", "service-uuid1", 8081)
		serviceInterface := new(mocks.ServiceInterface)
		serviceInterface.On("Create", expectedService).Return(nil, k8serrors.NewAlreadyExists(schema.GroupResource{}, ""))

		manager := NewAccessServiceManager(serviceInterface, config)

		// when
		err := manager.Upsert("ec-default", "appUID", "uuid-1", "service-uuid1")

		// then
		assert.NoError(t, err)
		serviceInterface.AssertExpectations(t)
	})

	t.Run("should handle errors", func(t *testing.T) {
		// given
		expectedService := mockService("ec-default", "uuid-1", "service-uuid1", 8081)
		serviceInterface := new(mocks.ServiceInterface)
		serviceInterface.On("Create", expectedService).Return(nil, errors.New("some error"))

		manager := NewAccessServiceManager(serviceInterface, config)

		// when
		err := manager.Upsert("ec-default", "appUID", "uuid-1", "service-uuid1")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceInterface.AssertExpectations(t)
	})
}

func TestAccessServiceManager_Delete(t *testing.T) {

	t.Run("should delete access service", func(t *testing.T) {
		// given
		serviceInterface := new(mocks.ServiceInterface)
		serviceInterface.On("Delete", "test-service", &metav1.DeleteOptions{}).Return(nil)

		manager := NewAccessServiceManager(serviceInterface, config)

		// when
		err := manager.Delete("test-service")

		// then
		assert.NoError(t, err)

		serviceInterface.AssertExpectations(t)
	})

	t.Run("should handle errors", func(t *testing.T) {
		// given
		serviceInterface := new(mocks.ServiceInterface)
		serviceInterface.On("Delete", "test-service", &metav1.DeleteOptions{}).Return(errors.New("some error"))

		manager := NewAccessServiceManager(serviceInterface, config)

		// when
		err := manager.Delete("test-service")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceInterface.AssertExpectations(t)
	})

	t.Run("should ignore not found error", func(t *testing.T) {
		// given
		serviceInterface := new(mocks.ServiceInterface)
		serviceInterface.On("Delete", "test-service", &metav1.DeleteOptions{}).Return(k8serrors.NewNotFound(schema.GroupResource{}, "an error"))

		manager := NewAccessServiceManager(serviceInterface, config)

		// when
		err := manager.Delete("test-service")

		// then
		assert.NoError(t, err)

		serviceInterface.AssertExpectations(t)
	})
}

func mockService(application, serviceId, serviceName string, targetPort int32) *corev1.Service {
	appName := fmt.Sprintf(appNameLabelFormat, application)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				k8sconsts.LabelApplication: application,
				k8sconsts.LabelServiceId:   serviceId,
			},
			OwnerReferences: k8sconsts.CreateOwnerReferenceForApplication(application, applicationUID),
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				k8sconsts.LabelApp: appName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: targetPort},
				},
			},
		},
	}
}
