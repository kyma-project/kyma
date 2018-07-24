package v1

import "k8s.io/apimachinery/pkg/util/intstr"

type Interface interface {
	Validate(name, namespace string, portNameOrNumber intstr.IntOrString) (ValidationResult, error)
}

type ValidationResult int

const (
	NotValidated ValidationResult = -1
	OK                            = iota
	FailedWithNotFound
	FailedWithInvalidPort
)
