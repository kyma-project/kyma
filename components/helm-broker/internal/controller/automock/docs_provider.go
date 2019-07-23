// Code generated by mockery v1.0.0
package automock

import internal "github.com/kyma-project/kyma/components/helm-broker/internal"
import mock "github.com/stretchr/testify/mock"

// DocsProvider is an autogenerated mock type for the DocsProvider type
type DocsProvider struct {
	mock.Mock
}

// EnsureDocsTopic provides a mock function with given fields: bundle, namespace
func (_m *DocsProvider) EnsureDocsTopic(bundle *internal.Bundle, namespace string) error {
	ret := _m.Called(bundle, namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(*internal.Bundle, string) error); ok {
		r0 = rf(bundle, namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnsureDocsTopicRemoved provides a mock function with given fields: id, namespace
func (_m *DocsProvider) EnsureDocsTopicRemoved(id string, namespace string) error {
	ret := _m.Called(id, namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(id, namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
