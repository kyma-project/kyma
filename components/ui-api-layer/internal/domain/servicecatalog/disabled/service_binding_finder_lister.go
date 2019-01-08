// Code generated by failery v1.0.0. DO NOT EDIT.

package disabled

import v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

// ServiceBindingFinderLister is an autogenerated failing mock type for the ServiceBindingFinderLister type
type ServiceBindingFinderLister struct {
	err error
}

// NewServiceBindingFinderLister creates a new ServiceBindingFinderLister type instance
func NewServiceBindingFinderLister(err error) *ServiceBindingFinderLister {
	return &ServiceBindingFinderLister{err: err}
}

// Find provides a failing mock function with given fields: env, name
func (_m *ServiceBindingFinderLister) Find(env string, name string) (*v1beta1.ServiceBinding, error) {
	var r0 *v1beta1.ServiceBinding
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ListForServiceInstance provides a failing mock function with given fields: env, instanceName
func (_m *ServiceBindingFinderLister) ListForServiceInstance(env string, instanceName string) ([]*v1beta1.ServiceBinding, error) {
	var r0 []*v1beta1.ServiceBinding
	var r1 error
	r1 = _m.err

	return r0, r1
}
