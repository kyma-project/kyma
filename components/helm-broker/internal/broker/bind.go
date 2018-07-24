package broker

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type bindService struct {
	instanceBindDataGetter instanceBindDataGetter
}

func (svc *bindService) Bind(ctx context.Context, osbCtx osbContext, req *osb.BindRequest) (*osb.BindResponse, error) {
	if len(req.Parameters) > 0 {
		return nil, fmt.Errorf("helm-broker does not support configuration options for the service binding")
	}

	out, err := svc.instanceBindDataGetter.Get(internal.InstanceID(req.InstanceID))
	if err != nil {
		return nil, errors.Wrapf(err, "while getting bind data from storage for instance id: %q", req.InstanceID)
	}

	return &osb.BindResponse{
		Credentials: svc.dtoFromModel(out.Credentials),
	}, nil
}

func (*bindService) dtoFromModel(in internal.InstanceCredentials) map[string]interface{} {
	dto := map[string]interface{}{}
	for k, v := range in {
		dto[k] = v
	}
	return dto
}
