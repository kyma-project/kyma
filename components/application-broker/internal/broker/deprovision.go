package broker

import (
	"context"
	"fmt"
	"sync"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/pkg/errors"
	"github.com/pmorie/go-open-service-broker-client/v2"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"
)

// NewDeprovisioner creates new Deprovisioner
func NewDeprovisioner(instStorage instanceStorage, instanceStateGetter instanceStateGetter, operationInserter operationInserter, operationUpdater operationUpdater, operationRemover operationRemover, opIDProvider func() (internal.OperationID, error), log logrus.FieldLogger) *DeprovisionService {
	return &DeprovisionService{
		instStorage:         instStorage,
		instanceStateGetter: instanceStateGetter,
		operationInserter:   operationInserter,
		operationUpdater:    operationUpdater,
		operationRemover:    operationRemover,
		operationIDProvider: opIDProvider,
		log:                 log.WithField("service", "deprovisioner"),
	}
}

// DeprovisionService performs deprovision action
type DeprovisionService struct {
	instStorage         instanceStorage
	instanceStateGetter instanceStateGetter
	operationIDProvider func() (internal.OperationID, error)
	operationInserter   operationInserter
	operationUpdater    operationUpdater
	operationRemover    operationRemover

	log       logrus.FieldLogger
	mu        sync.Mutex
	asyncHook func()
}

// Deprovision action
func (svc *DeprovisionService) Deprovision(ctx context.Context, osbCtx osbContext, req *v2.DeprovisionRequest) (*v2.DeprovisionResponse, error) {
	if !req.AcceptsIncomplete {
		return nil, errors.New("asynchronous operation mode required")
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	iID := internal.InstanceID(req.InstanceID)

	deprovisioned, err := svc.instanceStateGetter.IsDeprovisioned(iID)
	switch {
	case IsNotFoundError(err):
		return nil, err
	case err != nil:
		return nil, errors.Wrap(err, "while checking if instance is already deprovisioned")
	case deprovisioned:
		return &osb.DeprovisionResponse{Async: false}, nil
	}

	opIDInProgress, inProgress, err := svc.instanceStateGetter.IsDeprovisioningInProgress(iID)
	switch {
	case IsNotFoundError(err):
		return nil, err
	case err != nil:
		return nil, errors.Wrap(err, "while checking if instance is being deprovisioned")
	case inProgress:
		opKeyInProgress := osb.OperationKey(opIDInProgress)
		return &osb.DeprovisionResponse{Async: true, OperationKey: &opKeyInProgress}, nil
	}

	operationID, err := svc.operationIDProvider()
	if err != nil {
		return nil, errors.Wrap(err, "while generating ID for operation")
	}

	paramHash := "TODO"
	op := internal.InstanceOperation{
		InstanceID:  iID,
		OperationID: operationID,
		Type:        internal.OperationTypeRemove,
		State:       internal.OperationStateInProgress,
		ParamsHash:  paramHash,
	}

	if err := svc.operationInserter.Insert(&op); err != nil {
		return nil, errors.Wrap(err, "while inserting instance operation to storage")
	}

	err = svc.instStorage.Remove(iID)
	switch {
	case IsNotFoundError(err):
		return nil, err
	case err != nil:
		return nil, errors.Wrap(err, "while removing instance from storage")
	}

	opKey := osb.OperationKey(operationID)
	resp := &osb.DeprovisionResponse{
		Async:        true,
		OperationKey: &opKey,
	}

	svc.doAsync(iID, operationID, req.ServiceID)
	return resp, nil
}

func (svc *DeprovisionService) doAsync(iID internal.InstanceID, opID internal.OperationID, appID string) {
	go svc.do(iID, opID, appID)
}

func (svc *DeprovisionService) do(iID internal.InstanceID, opID internal.OperationID, appID string) {
	if svc.asyncHook != nil {
		defer svc.asyncHook()
	}

	opState := internal.OperationStateSucceeded
	opDesc := "deprovision succeeded"

	// remove instance entity from storage
	fDo := func() error {
		err := svc.operationRemover.Remove(iID, opID)
		switch {
		case err == nil, IsNotFoundError(err):
		default:
			return errors.Wrap(err, "while removing instance entity from storage")
		}
		return nil
	}

	if err := fDo(); err != nil {
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("deprovisioning failed on error: %s", err.Error())
	}

	if err := svc.operationUpdater.UpdateStateDesc(iID, opID, opState, &opDesc); err != nil {
		svc.log.Errorf("Cannot update state for instance [%s]: [%v]\n", iID, err)
		return
	}
}
