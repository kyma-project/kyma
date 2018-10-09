package accessservice

import (
	"fmt"
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/k8sconsts"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const appNameLabelFormat = "%s-gateway"

// ServiceInterface has methods to work with Service resources.
type ServiceInterface interface {
	Create(*corev1.Service) (*corev1.Service, error)
	Delete(name string, options *metav1.DeleteOptions) error
}

type AccessServiceManager interface {
	Create(remoteEnvironment, serviceId, serviceName string) apperrors.AppError
	Upsert(remoteEnvironment, serviceId, serviceName string) apperrors.AppError
	Delete(serviceName string) apperrors.AppError
}

type AccessServiceManagerConfig struct {
	TargetPort int32
}

type accessServiceManager struct {
	serviceInterface ServiceInterface
	config           AccessServiceManagerConfig
}

func NewAccessServiceManager(serviceInterface ServiceInterface, config AccessServiceManagerConfig) AccessServiceManager {
	return &accessServiceManager{
		serviceInterface: serviceInterface,
		config:           config,
	}
}

func (m *accessServiceManager) Create(remoteEnvironment, serviceId, serviceName string) apperrors.AppError {
	_, err := m.create(remoteEnvironment, serviceId, serviceName)
	if err != nil {
		return apperrors.Internal("Creating service failed, %s", err.Error())
	}
	return nil
}

func (m *accessServiceManager) Upsert(remoteEnvironment, serviceId, serviceName string) apperrors.AppError {
	_, err := m.create(remoteEnvironment, serviceId, serviceName)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}
		return apperrors.Internal("Upserting service failed, %s", err.Error())
	}
	return nil
}

func (m *accessServiceManager) Delete(serviceName string) apperrors.AppError {
	err := m.serviceInterface.Delete(serviceName, &metav1.DeleteOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return apperrors.Internal("Deleting service failed, %s", err.Error())
		}
	}
	return nil
}

func (m *accessServiceManager) create(remoteEnvironment, serviceId, serviceName string) (*corev1.Service, error) {
	appName := fmt.Sprintf(appNameLabelFormat, remoteEnvironment)

	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				k8sconsts.LabelRemoteEnvironment: remoteEnvironment,
				k8sconsts.LabelServiceId:         serviceId,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{k8sconsts.LabelApp: appName},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: m.config.TargetPort},
				},
			},
		},
	}

	return m.serviceInterface.Create(&service)
}
