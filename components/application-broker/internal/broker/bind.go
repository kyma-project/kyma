package broker

import (
	"context"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type getCredentialFn func(context.Context, string, internal.Service, string, string, string) (map[string]interface{}, error)

type bindService struct {
	appSvcFinder     appSvcFinder
	getCreds         getCredentialFn
	appSvcIDSelector appSvcIDSelector
}

const (
	fieldNameGatewayURL = "GATEWAY_URL"
)

func (svc *bindService) Bind(ctx context.Context, osbCtx osbContext, req *osb.BindRequest) (*osb.BindResponse, error) {
	if len(req.Parameters) > 0 {
		return nil, errors.New("application-broker does not support configuration options for the service binding")
	}

	appSvcID := svc.appSvcIDSelector.SelectID(req)
	app, err := svc.appSvcFinder.FindOneByServiceID(appSvcID)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get Application: %s", appSvcID)
	}

	// it is already validated, so it is safe to cast namespace
	ns := req.Context["namespace"].(string)

	creds, err := svc.getCredentials(ctx, ns, appSvcID, req.BindingID, req.InstanceID, app)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get credentials from applications")
	}

	return &osb.BindResponse{
		Credentials: creds,
	}, nil
}
func (svc *bindService) getCredentials(ctx context.Context, ns string, id internal.ApplicationServiceID, bindingID, instanceID string, app *internal.Application) (map[string]interface{}, error) {
	for idx := range app.Services {
		if app.Services[idx].ID == id {
			return svc.getCreds(ctx, ns, app.Services[idx], bindingID, app.CompassMetadata.ApplicationID, instanceID)
		}
	}

	return nil, errors.Errorf("cannot get credentials to bind instance with ApplicationServiceID: %s, from Application: %s", id, app.Name)
}
