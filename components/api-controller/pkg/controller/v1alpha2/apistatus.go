package v1alpha2

import (
	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/meta/v1"
	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/v1alpha2"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx/clientset/versioned"
	log "github.com/sirupsen/logrus"
)

type ApiStatusHelper struct {
	kymaInterface kyma.Interface
	apiCopy       *kymaApi.Api
	hasChanged    bool
}

func NewApiStatusHelper(kymaInterface kyma.Interface, api *kymaApi.Api) *ApiStatusHelper {
	return &ApiStatusHelper{
		kymaInterface: kymaInterface,
		apiCopy:       api.DeepCopy(),
		hasChanged:    false,
	}
}

func (su *ApiStatusHelper) SetAuthenticationStatus(authStatus *kymaMeta.GatewayResourceStatus) {
	su.apiCopy.Status.AuthenticationStatus = *authStatus
	su.hasChanged = true
}

func (su *ApiStatusHelper) SetVirtualServiceStatus(virtualServiceStatus *kymaMeta.GatewayResourceStatus) {
	su.apiCopy.Status.VirtualServiceStatus = *virtualServiceStatus
	su.hasChanged = true
}

func (su *ApiStatusHelper) Update() {

	if su.hasChanged {

		log.Infof("Saving status for: %s/%s ver: %s", su.apiCopy.Namespace, su.apiCopy.Name, su.apiCopy.ResourceVersion)
		log.Debugf("Status of %s/%s ver %s: %+v", su.apiCopy.Namespace, su.apiCopy.Name, su.apiCopy.ResourceVersion, su.apiCopy)

		if _, err2 := su.kymaInterface.GatewayV1alpha2().Apis(su.apiCopy.Namespace).Update(su.apiCopy); err2 != nil {
			log.Errorf("Error while saving API status. Root cause: %s", err2)
		}
	}
}
