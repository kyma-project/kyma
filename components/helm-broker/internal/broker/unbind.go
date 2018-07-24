package broker

import (
	"context"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type unbindService struct{}

func (svc *unbindService) Unbind(ctx context.Context, osbCtx osbContext, req *osb.UnbindRequest) (*osb.UnbindResponse, error) {
	return nil, nil
}
