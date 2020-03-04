package broker

import "github.com/kyma-project/kyma/components/application-broker/internal"

// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
func NewBindServiceV1(appFinder appSvcFinder) *bindService {
	return &bindService{appSvcFinder: appFinder, getCreds: getBindingCredentialsV1, appSvcIDSelector: &IDSelector{false}}
}

func NewBindServiceV2(appFinder appSvcFinder) *bindService {
	return &bindService{appSvcFinder: appFinder, getCreds: getBindingCredentialsV2, appSvcIDSelector: &IDSelector{true}}
}

func (svc *bindService) GetCredentials(id internal.ApplicationServiceID, app *internal.Application) (map[string]interface{}, error) {
	return svc.getCredentials(id, app)
}
