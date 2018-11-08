package broker

import (
	"context"

	"github.com/pkg/errors"

	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type getLastOperationService struct {
	getter operationGetter
}

func (svc *getLastOperationService) GetLastOperation(ctx context.Context, osbCtx OsbContext, req *osb.LastOperationRequest) (*osb.LastOperationResponse, error) {
	iID := internal.InstanceID(req.InstanceID)

	var opID internal.OperationID
	if req.OperationKey != nil {
		opID = internal.OperationID(*req.OperationKey)
	}

	op, err := svc.getter.Get(iID, opID)
	switch {
	case IsNotFoundError(err):
		return nil, err
	case err != nil:
		return nil, errors.Wrap(err, "while getting instance operation")
	}

	var descPtr *string
	if op.StateDescription != nil {
		desc := *op.StateDescription
		descPtr = &desc
	}

	resp := osb.LastOperationResponse{
		State:       osb.LastOperationState(op.State),
		Description: descPtr,
	}

	return &resp, nil
}
