// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"
import mock "github.com/stretchr/testify/mock"
import webhookconfig "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/webhookconfig"

// AssetWebhookConfigService is an autogenerated mock type for the AssetWebhookConfigService type
type AssetWebhookConfigService struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx
func (_m *AssetWebhookConfigService) Get(ctx context.Context) (webhookconfig.AssetWebhookConfigMap, error) {
	ret := _m.Called(ctx)

	var r0 webhookconfig.AssetWebhookConfigMap
	if rf, ok := ret.Get(0).(func(context.Context) webhookconfig.AssetWebhookConfigMap); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(webhookconfig.AssetWebhookConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
