package v1

import (
	"fmt"

	k8sCore "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8s "k8s.io/client-go/kubernetes"
)

type services struct {
	k8sInterface k8s.Interface
}

func New(k8sInterface k8s.Interface) Interface {
	return &services{
		k8sInterface: k8sInterface,
	}
}

func (s *services) Validate(name, namespace string, portNameOrNumber intstr.IntOrString) (ValidationResult, error) {

	service, err := s.get(name, namespace)
	if err != nil {
		return NotValidated, fmt.Errorf("Error while validating service: '%s'. Root cause:\n%v.", name, err)
	}
	return validateService(service, portNameOrNumber)
}

func (s *services) get(name, namespace string) (*k8sCore.Service, error) {

	servicesClient := s.k8sInterface.CoreV1().Services(namespace)

	getOptions := k8sMeta.GetOptions{}
	service, err := servicesClient.Get(name, getOptions)

	if err != nil {

		if apiErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Error while getting service: '%s'. Root cause:\n%v.", name, err)
	}

	return service, nil
}

func validateService(service *k8sCore.Service, portNameOrNumber intstr.IntOrString) (ValidationResult, error) {

	if service == nil {
		return FailedWithNotFound, nil
	}

	portValidationResult := validatePort(service, portNameOrNumber)
	if !portValidationResult {
		return FailedWithInvalidPort, nil
	}

	return OK, nil
}

func validatePort(service *k8sCore.Service, portNameOrNumber intstr.IntOrString) bool {

	validatePort := portValidatorFor(portNameOrNumber)

	ports := service.Spec.Ports
	for i := 0; i < len(ports); i++ {

		if validatePort(&ports[i]) {
			return true
		}
	}
	return false
}

type PortValidator func(port *k8sCore.ServicePort) bool

func portValidatorFor(portNameOrNumber intstr.IntOrString) PortValidator {

	switch portNameOrNumber.Type {
	case intstr.Int:
		portAsInt32 := portNameOrNumber.IntVal
		return func(port *k8sCore.ServicePort) bool {
			return port.Port == portAsInt32
		}
	case intstr.String:
		return func(port *k8sCore.ServicePort) bool {
			return port.Name == portNameOrNumber.StrVal
		}
	default:
		panic("Unsupported value type.")
	}
}
