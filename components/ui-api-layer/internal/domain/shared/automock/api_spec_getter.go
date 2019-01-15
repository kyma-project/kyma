// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

import storage "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"

// ApiSpecGetter is an autogenerated mock type for the ApiSpecGetter type
type ApiSpecGetter struct {
	mock.Mock
}

// Find provides a mock function with given fields: kind, id
func (_m *ApiSpecGetter) Find(kind string, id string) (*storage.ApiSpec, error) {
	ret := _m.Called(kind, id)

	var r0 *storage.ApiSpec
	if rf, ok := ret.Get(0).(func(string, string) *storage.ApiSpec); ok {
		r0 = rf(kind, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.ApiSpec)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(kind, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
