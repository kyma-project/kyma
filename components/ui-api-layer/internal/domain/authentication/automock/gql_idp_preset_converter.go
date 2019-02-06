// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"
import v1alpha1 "github.com/kyma-project/kyma/components/idppreset/pkg/apis/authentication/v1alpha1"

// gqlIDPPresetConverter is an autogenerated mock type for the gqlIDPPresetConverter type
type gqlIDPPresetConverter struct {
	mock.Mock
}

// ToGQL provides a mock function with given fields: in
func (_m *gqlIDPPresetConverter) ToGQL(in *v1alpha1.IDPPreset) gqlschema.IDPPreset {
	ret := _m.Called(in)

	var r0 gqlschema.IDPPreset
	if rf, ok := ret.Get(0).(func(*v1alpha1.IDPPreset) gqlschema.IDPPreset); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(gqlschema.IDPPreset)
	}

	return r0
}
