// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

import mock "github.com/stretchr/testify/mock"

// resourceQuotaStatusChecker is an autogenerated mock type for the resourceQuotaStatusChecker type
type resourceQuotaStatusChecker struct {
	mock.Mock
}

// CheckResourceQuotaStatus provides a mock function with given fields: namespace
func (_m *resourceQuotaStatusChecker) CheckResourceQuotaStatus(namespace string) (gqlschema.ResourceQuotasStatus, error) {
	ret := _m.Called(namespace)

	var r0 gqlschema.ResourceQuotasStatus
	if rf, ok := ret.Get(0).(func(string) gqlschema.ResourceQuotasStatus); ok {
		r0 = rf(namespace)
	} else {
		r0 = ret.Get(0).(gqlschema.ResourceQuotasStatus)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
