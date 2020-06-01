// Code generated by mockery v1.1.2. DO NOT EDIT.

package automock

import (
	internal "github.com/kyma-project/kyma/components/application-broker/internal"
	broker "github.com/kyma-project/kyma/components/application-broker/internal/broker"

	mock "github.com/stretchr/testify/mock"
)

// BrokerProcesses is an autogenerated mock type for the brokerProcesses type
type BrokerProcesses struct {
	mock.Mock
}

// DeprovisionProcess provides a mock function with given fields: _a0
func (_m *BrokerProcesses) DeprovisionProcess(_a0 broker.DeprovisionProcessRequest) {
	_m.Called(_a0)
}

// NewOperationID provides a mock function with given fields:
func (_m *BrokerProcesses) NewOperationID() (internal.OperationID, error) {
	ret := _m.Called()

	var r0 internal.OperationID
	if rf, ok := ret.Get(0).(func() internal.OperationID); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(internal.OperationID)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProvisionProcess provides a mock function with given fields: _a0
func (_m *BrokerProcesses) ProvisionProcess(_a0 broker.RestoreProvisionRequest) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(broker.RestoreProvisionRequest) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
