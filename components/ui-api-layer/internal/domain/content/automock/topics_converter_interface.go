// Code generated by mockery v1.0.0
package automock

import gqlschema "github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"
import storage "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"

// topicsConverterInterface is an autogenerated mock type for the topicsConverterInterface type
type topicsConverterInterface struct {
	mock.Mock
}

// ExtractSection provides a mock function with given fields: documents, internal
func (_m *topicsConverterInterface) ExtractSection(documents []storage.Document, internal bool) ([]gqlschema.Section, error) {
	ret := _m.Called(documents, internal)

	var r0 []gqlschema.Section
	if rf, ok := ret.Get(0).(func([]storage.Document, bool) []gqlschema.Section); ok {
		r0 = rf(documents, internal)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.Section)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]storage.Document, bool) error); ok {
		r1 = rf(documents, internal)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGQL provides a mock function with given fields: in
func (_m *topicsConverterInterface) ToGQL(in []gqlschema.TopicEntry) *gqlschema.JSON {
	ret := _m.Called(in)

	var r0 *gqlschema.JSON
	if rf, ok := ret.Get(0).(func([]gqlschema.TopicEntry) *gqlschema.JSON); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.JSON)
		}
	}

	return r0
}
