package automock

import (
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/mock"
)

func (_m *ApplicationCRMapper) ExpectOnToModel(dto *v1alpha1.Application, dm *internal.Application) *mock.Call {
	return _m.On("ToModel", dto).Return(dm, nil)
}

func (_m *ApplicationCRValidator) ExpectOnValidate(dto *v1alpha1.Application) *mock.Call {
	return _m.On("Validate", dto).Return(nil)
}

func (_m *ApplicationUpserter) ExpectOnUpsert(dm *internal.Application) *mock.Call {
	return _m.On("Upsert", dm).Return(false, nil)
}

func (_m *SCRelistRequester) ExpectOnRequestRelist() *mock.Call {
	return _m.On("RequestRelist")
}
