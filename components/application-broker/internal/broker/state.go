package broker

import (
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/application-broker/internal"
)

type instanceStateService struct {
	operationCollectionGetter operationCollectionGetter
}

func (svc *instanceStateService) IsProvisioned(iID internal.InstanceID) (bool, error) {
	result := false

	ops, err := svc.operationCollectionGetter.GetAll(iID)
	switch {
	case err == nil:
	case IsNotFoundError(err):
		return false, nil
	default:
		return false, errors.Wrap(err, "while getting operations from storage")
	}

OpsLoop:
	for _, op := range ops {
		if op.Type == internal.OperationTypeCreate && op.State == internal.OperationStateSucceeded {
			result = true
		}
		if op.Type == internal.OperationTypeRemove && op.State == internal.OperationStateSucceeded {
			result = false
			break OpsLoop
		}
	}

	return result, nil
}

func (svc *instanceStateService) IsProvisioningInProgress(iID internal.InstanceID) (internal.OperationID, bool, error) {
	resultInProgress := false
	var resultOpID internal.OperationID

	ops, err := svc.operationCollectionGetter.GetAll(iID)
	switch {
	case err == nil:
	case IsNotFoundError(err):
		return resultOpID, false, nil
	default:
		return resultOpID, false, errors.Wrap(err, "while getting operations from storage")
	}

OpsLoop:
	for _, op := range ops {
		if op.Type == internal.OperationTypeCreate && op.State == internal.OperationStateInProgress {
			resultInProgress = true
			resultOpID = op.OperationID
			break OpsLoop
		}
	}

	return resultOpID, resultInProgress, nil
}

func (svc *instanceStateService) IsDeprovisioned(iID internal.InstanceID) (bool, error) {
	result := false

	ops, err := svc.operationCollectionGetter.GetAll(iID)
	switch {
	case err == nil:
	case IsNotFoundError(err):
		return false, err
	default:
		return false, errors.Wrap(err, "while getting operations from storage")
	}

OpsLoop:
	for _, op := range ops {
		if op.Type == internal.OperationTypeRemove && op.State == internal.OperationStateSucceeded {
			result = true
			break OpsLoop
		}
	}

	return result, nil
}

func (svc *instanceStateService) IsDeprovisioningInProgress(iID internal.InstanceID) (internal.OperationID, bool, error) {
	resultInProgress := false
	var resultOpID internal.OperationID

	ops, err := svc.operationCollectionGetter.GetAll(iID)
	switch {
	case err == nil:
	case IsNotFoundError(err):
		return resultOpID, false, nil
	default:
		return resultOpID, false, errors.Wrap(err, "while getting operations from storage")
	}

OpsLoop:
	for _, op := range ops {
		if op.Type == internal.OperationTypeRemove && op.State == internal.OperationStateInProgress {
			resultInProgress = true
			resultOpID = op.OperationID
			break OpsLoop
		}
	}

	return resultOpID, resultInProgress, nil
}
