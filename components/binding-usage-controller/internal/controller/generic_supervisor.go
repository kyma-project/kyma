package controller

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// InterfaceProvider provides dynamic.ResourceInterface for given namespace
type InterfaceProvider interface {
	ResourceInterface(namespace string) dynamic.ResourceInterface
}

type labelDeleteApplier interface {
	EnsureLabelsAreApplied(res *unstructured.Unstructured, labels map[string]string) error
	EnsureLabelsAreDeleted(res *unstructured.Unstructured, labels map[string]string) error
}

// GenericSupervisor ensures that expected labels are present or not on a k8s resource provided by given InterfaceProvider.
type GenericSupervisor struct {
	log logrus.FieldLogger

	interfaceProvider InterfaceProvider
	tracer            genericUsageBindingAnnotationTracer
	labelSvc          labelDeleteApplier
}

// NewGenericSupervisor creates a new GenericSupervisor.
func NewGenericSupervisor(interfaceProvider InterfaceProvider, labeler labelDeleteApplier, log logrus.FieldLogger) *GenericSupervisor {
	return &GenericSupervisor{
		log: log,

		tracer:            &genericUsageAnnotationTracer{},
		interfaceProvider: interfaceProvider,
		labelSvc:          labeler,
	}
}

// EnsureLabelsCreated ensures that given labels are added to resource
func (m *GenericSupervisor) EnsureLabelsCreated(namespace, resourceName, usageName string, labels map[string]string) error {
	res, err := m.interfaceProvider.ResourceInterface(namespace).Get(resourceName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "while getting resource")
	}

	// TODO (pluggable SBU cleanup): implement detecting labels conflict, see deployment_supervisor.go

	// apply new labels
	if err := m.labelSvc.EnsureLabelsAreApplied(res, labels); err != nil {
		return errors.Wrap(err, "while ensuring labels are applied")
	}

	if err := m.tracer.SetAnnotationAboutBindingUsage(res, usageName, labels); err != nil {
		return errors.Wrap(err, "while setting annotation tracing data")
	}

	if err := m.executeUpdate(res); err != nil {
		return errors.Wrap(err, "while updating resource")
	}

	return nil
}

// EnsureLabelsDeleted ensures that given labels are deleted on resource
func (m *GenericSupervisor) EnsureLabelsDeleted(namespace, resourceName, usageName string) error {
	res, err := m.interfaceProvider.ResourceInterface(namespace).Get(resourceName, metav1.GetOptions{})
	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		return nil
	default:
		return errors.Wrap(err, "while getting resource")
	}

	labels, err := m.tracer.GetInjectedLabels(res, usageName)
	if err != nil {
		return errors.Wrap(err, "while getting injected labels keys")
	}

	// remove old labels
	if err := m.labelSvc.EnsureLabelsAreDeleted(res, labels); err != nil {
		return errors.Wrap(err, "while getting injected labels keys")
	}

	// remove annotations
	err = m.tracer.DeleteAnnotationAboutBindingUsage(res, usageName)
	if err != nil {
		return errors.Wrap(err, "while deleting annotation tracing data")
	}

	if err := m.executeUpdate(res); err != nil {
		return errors.Wrap(err, "while updating resource")
	}

	return nil
}

// GetInjectedLabels returns labels applied on given resource by usage controller
func (m *GenericSupervisor) GetInjectedLabels(namespace, resourceName, usageName string) (map[string]string, error) {
	res, err := m.interfaceProvider.ResourceInterface(namespace).Get(resourceName, metav1.GetOptions{})
	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		return nil, NewNotFoundError(err.Error())
	default:
		return nil, errors.Wrap(err, "while listing resources")
	}

	labels, err := m.tracer.GetInjectedLabels(res, usageName)
	if err != nil {
		return nil, errors.Wrap(err, "while getting injected labels keys")
	}

	return labels, nil
}

// HasSynced returns true, because the GenericSupervisor does not use any informers/caches
func (m *GenericSupervisor) HasSynced() bool {
	return true
}

func (m *GenericSupervisor) executeUpdate(res *unstructured.Unstructured) error {
	_, err := m.interfaceProvider.ResourceInterface(res.GetNamespace()).Update(res)
	if err != nil {
		return errors.Wrapf(err, "while updating %s %s in namespace %s", res.GetKind(), res.GetName(), res.GetNamespace())
	}

	return nil
}
