package broker

import (
	"context"
	"sync"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"
)

// NewDeprovisioner creates new Deprovisioner
func NewDeprovisioner(instStorage instanceStorage, instanceStateGetter instanceStateGetter, operationInserter operationInserter, operationUpdater operationUpdater, opIDProvider func() (internal.OperationID, error), log logrus.FieldLogger) *DeprovisionService {
	return &DeprovisionService{
		instStorage:         instStorage,
		instanceStateGetter: instanceStateGetter,
		operationInserter:   operationInserter,
		operationUpdater:    operationUpdater,
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

	log       logrus.FieldLogger
	mu        sync.Mutex
	asyncHook func()
}

// Deprovision action
func (svc *DeprovisionService) Deprovision(ctx context.Context, osbCtx osbContext, req *osb.DeprovisionRequest) (*osb.DeprovisionResponse, error) {
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

	iNs, err := svc.instStorage.Get(iID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting instance from storage")
	}

	op := internal.InstanceOperation{
		InstanceID:  iID,
		OperationID: operationID,
		Type:        internal.OperationTypeRemove,
		State:       internal.OperationStateInProgress,
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

	svc.doAsync(iID, operationID, req.ServiceID, iNs.Namespace)
	return resp, nil
}

func (svc *DeprovisionService) doAsync(iID internal.InstanceID, opID internal.OperationID, appID string, ns internal.Namespace) {
	go svc.do(iID, opID, appID, ns)
}

func (svc *DeprovisionService) do(iID internal.InstanceID, opID internal.OperationID, appID string, ns internal.Namespace) {
	if svc.asyncHook != nil {
		defer svc.asyncHook()
	}

	opState := internal.OperationStateSucceeded
	opDesc := "deprovision succeeded"

	// currently, there is no any action, but it is a place for future - any deprovisioning action should be put here

	if err := svc.operationUpdater.UpdateStateDesc(iID, opID, opState, &opDesc); err != nil {
		svc.log.Errorf("Cannot update state for instance [%s]: [%v]\n", iID, err)
		return
	}
}
