package automock

import (
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/mock"
)

func (_m *RemoteEnvironmentCRMapper) ExpectOnToModel(dto *v1alpha1.RemoteEnvironment, dm *internal.RemoteEnvironment) *mock.Call {
	return _m.On("ToModel", dto).Return(dm)
}

func (_m *RemoteEnvironmentCRValidator) ExpectOnValidate(dto *v1alpha1.RemoteEnvironment) *mock.Call {
	return _m.On("Validate", dto).Return(nil)
}

func (_m *RemoteEnvironmentUpserter) ExpectOnUpsert(dm *internal.RemoteEnvironment) *mock.Call {
	return _m.On("Upsert", dm).Return(false, nil)
}

func (_m *SCRelistRequester) ExpectOnRequestRelist() *mock.Call {
	return _m.On("RequestRelist")
}
