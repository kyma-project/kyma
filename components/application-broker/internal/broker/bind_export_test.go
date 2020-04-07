package broker

import (
	"context"

	"github.com/kyma-project/kyma/components/application-broker/internal"
)

// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
func NewBindServiceV1(appFinder appSvcFinder) *bindService {
	renderer := BindingCredentialsRenderer{}
	return &bindService{appSvcFinder: appFinder, getCreds: renderer.GetBindingCredentialsV1, appSvcIDSelector: &IDSelector{false}}
}

func NewBindServiceV2(appFinder appSvcFinder, apiPkgCredGetter apiPackageCredentialsGetter, gatewayBaseURLFormat string, sbFetcher ServiceBindingFetcher) *bindService {
	renderer := BindingCredentialsRenderer{
		apiPackageCredGetter: apiPkgCredGetter,
		gatewayBaseURLFormat: gatewayBaseURLFormat,
		sbFetcher:            sbFetcher,
	}
	return &bindService{appSvcFinder: appFinder, getCreds: renderer.GetBindingCredentialsV2, appSvcIDSelector: &IDSelector{true}}
}

func (svc *bindService) GetCredentials(ctx context.Context, namespace string, appSvcID internal.ApplicationServiceID, bindingID string, instanceID string, app *internal.Application) (map[string]interface{}, error) {
	return svc.getCredentials(ctx, namespace, appSvcID, bindingID, instanceID, app)
}
