// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

import mock "github.com/stretchr/testify/mock"
import v1 "k8s.io/api/core/v1"

// namespaceConverter is an autogenerated mock type for the namespaceConverter type
type namespaceConverter struct {
	mock.Mock
}

// ToListItemGQL provides a mock function with given fields: in
func (_m *namespaceConverter) ToListItemGQL(in *v1.Namespace) *gqlschema.NamespaceListItem {
	ret := _m.Called(in)

	var r0 *gqlschema.NamespaceListItem
	if rf, ok := ret.Get(0).(func(*v1.Namespace) *gqlschema.NamespaceListItem); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.NamespaceListItem)
		}
	}

	return r0
}
