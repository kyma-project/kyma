// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	mock "github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ResourceGetter is an autogenerated mock type for the ResourceGetter type
type ResourceGetter struct {
	mock.Mock
}

// Get provides a mock function with given fields: name, options, subresources
func (_m *ResourceGetter) Get(name string, options v1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	_va := make([]interface{}, len(subresources))
	for _i := range subresources {
		_va[_i] = subresources[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, options)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *unstructured.Unstructured
	if rf, ok := ret.Get(0).(func(string, v1.GetOptions, ...string) *unstructured.Unstructured); ok {
		r0 = rf(name, options, subresources...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*unstructured.Unstructured)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, v1.GetOptions, ...string) error); ok {
		r1 = rf(name, options, subresources...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
