package broker

import "github.com/kyma-project/kyma/components/application-broker/internal"

func NewBindService(appFinder appFinder) *bindService {
	return &bindService{
		reSvcFinder: appFinder,
	}
}

func (svc *bindService) GetCredentials(id internal.ApplicationServiceID, app *internal.Application) (map[string]interface{}, error) {
	return svc.getCredentials(id, app)
}
