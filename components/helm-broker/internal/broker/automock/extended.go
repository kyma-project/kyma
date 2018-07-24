package automock

import (
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/stretchr/testify/mock"
)

// InstanceStateGetter extensions
func (_m *InstanceStateGetter) ExpectOnIsDeprovisioned(iID internal.InstanceID, deprovisioned bool) *mock.Call {
	return _m.On("IsDeprovisioned", iID).Return(deprovisioned, nil)
}

func (_m *InstanceStateGetter) ExpectOnIsDeprovisioningInProgress(iID internal.InstanceID, optID internal.OperationID, inProgress bool) *mock.Call {
	return _m.On("IsDeprovisioningInProgress", iID).Return(optID, inProgress, nil)
}

func (_m *InstanceStateGetter) ExpectErrorOnIsDeprovisioningInProgress(iID internal.InstanceID, err error) *mock.Call {
	return _m.On("IsDeprovisioningInProgress", iID).Return(internal.OperationID(""), false, err)
}

func (_m *InstanceStateGetter) ExpectErrorIsDeprovisioned(iID internal.InstanceID, err error) *mock.Call {
	return _m.On("IsDeprovisioned", iID).Return(false, err)
}

// InstanceStorage extensions
func (_m *InstanceStorage) ExpectOnGet(iID internal.InstanceID, expInstance internal.Instance) *mock.Call {
	return _m.On("Get", iID).Return(&expInstance, nil)
}

func (_m *InstanceStorage) ExpectErrorOnGet(iID internal.InstanceID, err error) *mock.Call {
	return _m.On("Get", iID).Return(nil, err)
}

// OperationStorage extensions
func (_m *OperationStorage) ExpectOnInsert(op internal.InstanceOperation) *mock.Call {
	return _m.On("Insert", &op).Return(nil)
}

func (_m *OperationStorage) ExpectOnUpdateStateDesc(iID internal.InstanceID, opID internal.OperationID, state internal.OperationState, desc string) *mock.Call {
	return _m.On("UpdateStateDesc", iID, opID, state, &desc).Return(nil)
}

// HelmClient extensions
func (_m *HelmClient) ExpectOnDelete(rName internal.ReleaseName) *mock.Call {
	return _m.On("Delete", rName).Return(nil)
}

func (_m *HelmClient) ExpectErrorOnDelete(rName internal.ReleaseName, err error) *mock.Call {
	return _m.On("Delete", rName).Return(err)
}

// InstanceBindDataRemover extensions
func (_m *InstanceBindDataRemover) ExpectOnRemove(iID internal.InstanceID) *mock.Call {
	return _m.On("Remove", iID).Return(nil)
}

func (_m *InstanceBindDataRemover) ExpectErrorRemove(iID internal.InstanceID, err error) *mock.Call {
	return _m.On("Remove", iID).Return(err)
}
