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
		return nil, errors.New("application-broker does not support configuration options for the service binding")
	}

	app, err := svc.reSvcFinder.FindOneByServiceID(internal.ApplicationServiceID(req.ServiceID))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get Application: %s", req.ServiceID)
	}

	creds, err := svc.getCredentials(internal.ApplicationServiceID(req.ServiceID), app)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get credentials from applications")
	}

	return &osb.BindResponse{
		Credentials: creds,
	}, nil
}

func (*bindService) getCredentials(rsID internal.ApplicationServiceID, app *internal.Application) (map[string]interface{}, error) {
	creds := make(map[string]interface{})
	for _, svc := range app.Services {
		if svc.ID == rsID {
			creds[fieldNameGatewayURL] = svc.APIEntry.GatewayURL
			return creds, nil
		}
	}
	return nil, errors.Errorf("cannot get credentials to bind instance with ApplicationServiceID: %s, from Application: %s", rsID, app.Name)
}
