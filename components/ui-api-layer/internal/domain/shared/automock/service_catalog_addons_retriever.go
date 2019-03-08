// Code generated by mockery v1.0.0. DO NOT EDIT.
package automock

import mock "github.com/stretchr/testify/mock"
import shared "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/shared"

// ServiceCatalogAddonsRetriever is an autogenerated mock type for the ServiceCatalogAddonsRetriever type
type ServiceCatalogAddonsRetriever struct {
	mock.Mock
}

// ServiceBindingUsage provides a mock function with given fields:
func (_m *ServiceCatalogAddonsRetriever) ServiceBindingUsage() shared.ServiceBindingUsageLister {
	ret := _m.Called()

	var r0 shared.ServiceBindingUsageLister
	if rf, ok := ret.Get(0).(func() shared.ServiceBindingUsageLister); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(shared.ServiceBindingUsageLister)
		}
	}

	return r0
}
