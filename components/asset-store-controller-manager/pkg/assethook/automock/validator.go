// Code generated by mockery v1.0.0
package automock

import assethook "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook"
import context "context"
import mock "github.com/stretchr/testify/mock"
import v1alpha2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"

// Validator is an autogenerated mock type for the Validator type
type Validator struct {
	mock.Mock
}

// Validate provides a mock function with given fields: ctx, basePath, files, services
func (_m *Validator) Validate(ctx context.Context, basePath string, files []string, services []v1alpha2.AssetWebhookService) (assethook.Result, error) {
	ret := _m.Called(ctx, basePath, files, services)

	var r0 assethook.Result
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, []v1alpha2.AssetWebhookService) assethook.Result); ok {
		r0 = rf(ctx, basePath, files, services)
	} else {
		r0 = ret.Get(0).(assethook.Result)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string, []v1alpha2.AssetWebhookService) error); ok {
		r1 = rf(ctx, basePath, files, services)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
