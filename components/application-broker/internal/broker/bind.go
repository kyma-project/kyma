package broker

import (
	"context"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type bindService struct {
	reSvcFinder reSvcFinder
}

const fieldNameGatewayURL = "GATEWAY_URL"

func (svc *bindService) Bind(ctx context.Context, osbCtx osbContext, req *osb.BindRequest) (*osb.BindResponse, error) {
	if len(req.Parameters) > 0 {
		return nil, errors.New("remote-environment-broker does not support configuration options for the service binding")
	}

	re, err := svc.reSvcFinder.FindOneByServiceID(internal.RemoteServiceID(req.ServiceID))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get RemoteEnvironment: %s", req.ServiceID)
	}

	creds, err := svc.getCredentials(internal.RemoteServiceID(req.ServiceID), re)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get credentials from remote environments")
	}

	return &osb.BindResponse{
		Credentials: creds,
	}, nil
}

func (*bindService) getCredentials(rsID internal.RemoteServiceID, re *internal.RemoteEnvironment) (map[string]interface{}, error) {
	creds := make(map[string]interface{})
	for _, svc := range re.Services {
		if svc.ID == rsID {
			creds[fieldNameGatewayURL] = svc.APIEntry.GatewayURL
			return creds, nil
		}
	}
	return nil, errors.Errorf("cannot get credentials to bind instance with RemoteServiceID: %s, from RemoteEnvironment: %s", rsID, re.Name)
}
