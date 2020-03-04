package broker

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type bindService struct {
	appSvcFinder     appSvcFinder
	getCreds         func([]internal.Entry) map[string]interface{}
	appSvcIDSelector AppSvcIDSelector
}

const (
	fieldNameGatewayURL    = "GATEWAY_URL"
	fieldPatternGatewayURL = "%s_GATEWAY_URL"
	fieldPatternTargetURL  = "%s_TARGET_URL"
)

func (svc *bindService) Bind(ctx context.Context, osbCtx osbContext, req *osb.BindRequest) (*osb.BindResponse, error) {
	if len(req.Parameters) > 0 {
		return nil, errors.New("application-broker does not support configuration options for the service binding")
	}

	id := svc.appSvcIDSelector.SelectID(req)
	app, err := svc.appSvcFinder.FindOneByServiceID(id)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get Application: %s", id)
	}

	creds, err := svc.getCredentials(id, app)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get credentials from applications")
	}

	return &osb.BindResponse{
		Credentials: creds,
	}, nil
}

func (svc *bindService) getCredentials(id internal.ApplicationServiceID, app *internal.Application) (map[string]interface{}, error) {
	for _, s := range app.Services {
		if s.ID == id {
			return svc.getCreds(s.Entries), nil
		}
	}

	return nil, errors.Errorf("cannot get credentials to bind instance with ApplicationServiceID: %s, from Application: %s", id, app.Name)
}

func getBindingCredentialsV2(entries []internal.Entry) map[string]interface{} {
	creds := make(map[string]interface{})
	for _, e := range entries {
		if e.Type == internal.APIEntryType {
			creds[strings.ToUpper(fmt.Sprintf(fieldPatternGatewayURL, e.Name))] = e.GatewayURL
			creds[strings.ToUpper(fmt.Sprintf(fieldPatternTargetURL, e.Name))] = e.TargetURL
		}
	}
	return creds
}

// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
// in old approach if it is bindable then it has only one API entry
func getBindingCredentialsV1(entries []internal.Entry) map[string]interface{} {
	creds := make(map[string]interface{})
	creds[fieldNameGatewayURL] = entries[0].GatewayURL
	return creds
}
