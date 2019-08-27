package v1

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestValidate_ShouldReturnNotFound(t *testing.T) {

	validationResult, err := validateService(nil, intstr.FromString("http"))

	if err != nil {
		t.Errorf("Error validating service. Root cause:\n%v", err)
	}

	if validationResult != FailedWithNotFound {
		t.Errorf("Should return: %v but got: %v", FailedWithNotFound, validationResult)
	}
}

func TestValidate_ShouldReturnInvalidPort(t *testing.T) {

	service := v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Name: "http"},
			},
		},
	}

	validationResult, err := validateService(&service, intstr.FromInt(80))

	if err != nil {
		t.Errorf("Error validating service. Root cause:\n%v", err)
	}

	if validationResult != FailedWithInvalidPort {
		t.Errorf("Should return: %v but got: %v", FailedWithInvalidPort, validationResult)
	}
}

func TestValidate_ShouldReturnOK(t *testing.T) {

	service := v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: 80},
			},
		},
	}

	validationResult, err := validateService(&service, intstr.FromInt(80))

	if err != nil {
		t.Errorf("Error validating service. Root cause:\n%v", err)
	}

	if validationResult != OK {
		t.Errorf("Should return: %v but got: %v", OK, validationResult)
	}
}

func TestValidatePort_ShouldReturnTrueForPortName(t *testing.T) {

	service := v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Name: "http"},
			},
		},
	}

	result := validatePort(&service, intstr.FromString("http"))

	if !result {
		t.Errorf("Should return true for service port name")
	}
}

func TestValidatePort_ShouldReturnFalseForPortName(t *testing.T) {

	service := v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Name: "http"},
			},
		},
	}

	result := validatePort(&service, intstr.FromString("http-but-invalid"))

	if result {
		t.Errorf("Should return false for service port name")
	}
}

func TestValidatePort_ShouldReturnTrueForPortNumber(t *testing.T) {

	service := v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: 80},
			},
		},
	}

	result := validatePort(&service, intstr.FromInt(80))

	if !result {
		t.Errorf("Should return true for service port name")
	}
}

func TestValidatePort_ShouldReturnFalseForPortNumber(t *testing.T) {

	service := v1.Service{
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: 80},
			},
		},
	}

	result := validatePort(&service, intstr.FromInt(81))

	if result {
		t.Errorf("Should return false for service port name")
	}
}
