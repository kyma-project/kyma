// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"
import mock "github.com/stretchr/testify/mock"
import v1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
import webhook "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/asset/webhook"

// Validator is an autogenerated mock type for the Validator type
type Validator struct {
	mock.Mock
}

// Validate provides a mock function with given fields: ctx, basePath, files, asset
func (_m *Validator) Validate(ctx context.Context, basePath string, files []string, asset *v1alpha1.Asset) (webhook.ValidationResult, error) {
	ret := _m.Called(ctx, basePath, files, asset)

	var r0 webhook.ValidationResult
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, *v1alpha1.Asset) webhook.ValidationResult); ok {
		r0 = rf(ctx, basePath, files, asset)
	} else {
		r0 = ret.Get(0).(webhook.ValidationResult)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string, *v1alpha1.Asset) error); ok {
		r1 = rf(ctx, basePath, files, asset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
