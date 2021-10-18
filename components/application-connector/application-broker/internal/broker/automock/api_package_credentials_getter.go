// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"
import internal "github.com/kyma-project/kyma/components/application-connector/application-broker/internal"
import mock "github.com/stretchr/testify/mock"

// apiPackageCredentialsGetter is an autogenerated mock type for the apiPackageCredentialsGetter type
type apiPackageCredentialsGetter struct {
	mock.Mock
}

// GetAPIPackageCredentials provides a mock function with given fields: ctx, appID, pkgID, instanceID
func (_m *apiPackageCredentialsGetter) GetAPIPackageCredentials(ctx context.Context, appID string, pkgID string, instanceID string) (internal.APIPackageCredential, error) {
	ret := _m.Called(ctx, appID, pkgID, instanceID)

	var r0 internal.APIPackageCredential
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) internal.APIPackageCredential); ok {
		r0 = rf(ctx, appID, pkgID, instanceID)
	} else {
		r0 = ret.Get(0).(internal.APIPackageCredential)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, appID, pkgID, instanceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
