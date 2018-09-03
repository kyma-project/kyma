// Code generated by mockery v1.0.0
package automock

import mock "github.com/stretchr/testify/mock"
import types "k8s.io/apimachinery/pkg/types"
import v1 "k8s.io/api/core/v1"

// NsPatcher is an autogenerated mock type for the NsPatcher type
type NsPatcher struct {
	mock.Mock
}

// Patch provides a mock function with given fields: name, pt, data, subresources
func (_m *NsPatcher) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v1.Namespace, error) {
	_va := make([]interface{}, len(subresources))
	for _i := range subresources {
		_va[_i] = subresources[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, pt, data)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *v1.Namespace
	if rf, ok := ret.Get(0).(func(string, types.PatchType, []byte, ...string) *v1.Namespace); ok {
		r0 = rf(name, pt, data, subresources...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Namespace)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, types.PatchType, []byte, ...string) error); ok {
		r1 = rf(name, pt, data, subresources...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
