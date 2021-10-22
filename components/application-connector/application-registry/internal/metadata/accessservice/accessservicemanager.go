package accessservice

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const appNameLabelFormat = "%s-application-gateway"

// ServiceInterface has methods to work with Service resources.
//go:generate mockery --name ServiceInterface
type ServiceInterface interface {
	Create(context.Context, *corev1.Service, metav1.CreateOptions) (*corev1.Service, error)
	Delete(context.Context, string, metav1.DeleteOptions) error
}

type AccessServiceManager interface {
	Create(application string, appUID types.UID, serviceId, serviceName string) apperrors.AppError
	Upsert(application string, appUID types.UID, serviceId, serviceName string) apperrors.AppError
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

func (m *accessServiceManager) Create(application string, appUID types.UID, serviceId, serviceName string) apperrors.AppError {
	_, err := m.create(application, appUID, serviceId, serviceName)
	if err != nil {
		return apperrors.Internal("Creating service failed, %s", err.Error())
	}
	return nil
}

func (m *accessServiceManager) Upsert(application string, appUID types.UID, serviceId, serviceName string) apperrors.AppError {
	_, err := m.create(application, appUID, serviceId, serviceName)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}
		return apperrors.Internal("Upserting service failed, %s", err.Error())
	}
	return nil
}

func (m *accessServiceManager) Delete(serviceName string) apperrors.AppError {
	err := m.serviceInterface.Delete(context.Background(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return apperrors.Internal("Deleting service failed, %s", err.Error())
		}
	}
	return nil
}

func (m *accessServiceManager) create(application string, appUID types.UID, serviceId, serviceName string) (*corev1.Service, error) {
	appName := fmt.Sprintf(appNameLabelFormat, application)

	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				k8sconsts.LabelApplication: application,
				k8sconsts.LabelServiceId:   serviceId,
			},
			OwnerReferences: k8sconsts.CreateOwnerReferenceForApplication(application, appUID),
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

	return m.serviceInterface.Create(context.Background(), &service, metav1.CreateOptions{})
}
