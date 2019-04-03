package v2

import (
	"reflect"

	istioAuthApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/authentication.istio.io/v1alpha1"
	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/meta/v1"
	istioAuth "github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io/clientset/versioned"
	istioAuthTyped "github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io/clientset/versioned/typed/authentication.istio.io/v1alpha1"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/commons"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	log "github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type istioImpl struct {
	istioAuthInterface        istioAuth.Interface
	jwtDefaultConfig          JwtDefaultConfig
	enableIstioAuthPolicyMTLS bool
}

func New(a istioAuth.Interface, c JwtDefaultConfig, enableIstioAuthPolicyMTLS bool) Interface {
	return &istioImpl{
		istioAuthInterface:        a,
		jwtDefaultConfig:          c,
		enableIstioAuthPolicyMTLS: enableIstioAuthPolicyMTLS,
	}
}

func (a *istioImpl) Create(dto *Dto) (*kymaMeta.GatewayResource, error) {

	if isAuthenticationDisabled(dto) {
		return nil, nil
	}

	istioAuthPolicy := toIstioAuthPolicy(dto, a.jwtDefaultConfig, a.enableIstioAuthPolicyMTLS)

	log.Infof("Creating authentication policy: %+v", istioAuthPolicy)

	created, err := a.istioAuthPolicyInterface(dto.MetaDto).Create(istioAuthPolicy)
	if err != nil {
		return nil, commons.HandleError(err, "error while creating authentication policy")
	}

	log.Infof("Authentication policy %s/%s ver: %s created.", istioAuthPolicy.Namespace, istioAuthPolicy.Name, istioAuthPolicy.ResourceVersion)
	return gatewayResourceFrom(created), nil
}

func (a *istioImpl) Update(oldDto, newDto *Dto) (*kymaMeta.GatewayResource, error) {

	if isAuthenticationDisabled(newDto) {

		log.Infof("Authentication disabled. Trying to delete the old authentication policy...")
		// no new new policy; we only have to delete the old one, if it exists
		if err := a.Delete(oldDto); err != nil {
			return nil, err
		}
		return nil, nil
	}

	// there is an authentication policy to update / create
	newIstioAuthPolicy := toIstioAuthPolicy(newDto, a.jwtDefaultConfig, a.enableIstioAuthPolicyMTLS)

	log.Infof("Authentication enabled. Trying to create or update authentication policy with: %v", newIstioAuthPolicy)

	// checking if authentication policy has to be created (was disabled before)
	if isAuthenticationDisabled(oldDto) || oldDto.Status.Resource.Name == "" {

		log.Infof("Authentication policy does not exist. Creating...")

		// create new authentication policy
		createdResource, err := a.Create(newDto)

		if err != nil {
			return nil, commons.HandleError(err, "error while creating authentication policy (can not create a new one)")
		}

		log.Infof("Authentication policy created: %v", createdResource)
		return createdResource, nil
	}

	oldIstioAuthPolicy := toIstioAuthPolicy(oldDto, a.jwtDefaultConfig, a.enableIstioAuthPolicyMTLS)

	if a.isEqual(oldIstioAuthPolicy, newIstioAuthPolicy) {

		log.Infof("Update skipped: authentication policy has not changed.")
		return gatewayResourceFrom(oldIstioAuthPolicy), nil
	}

	newIstioAuthPolicy.ObjectMeta.ResourceVersion = oldDto.Status.Resource.Version

	// new authentication policy should be updated (it was created earlier and it differs from the old one)
	log.Infof("Updating authentication policy: %v", newIstioAuthPolicy)

	updated, err := a.istioAuthPolicyInterface(newDto.MetaDto).Update(newIstioAuthPolicy)
	if err != nil {
		return nil, commons.HandleError(err, "error while updating authentication policy")
	}

	log.Infof("Authentication policy updated: %v", updated)
	return gatewayResourceFrom(updated), nil
}

func (a *istioImpl) Delete(dto *Dto) error {

	if isAuthenticationDisabled(dto) || dto.Status.Resource.Name == "" {
		log.Infof("Delete skipped: no authentication policy to delete.")
		return nil
	}
	return a.deleteByName(dto.MetaDto)
}

func (a *istioImpl) deleteByName(meta meta.Dto) error {

	// if there is no authentication policy to delete, just skip it
	if meta.Name == "" {
		log.Infof("Delete skipped: no authentication policy to delete.")
		return nil
	}
	log.Infof("Deleting authentication policy: %s", meta.Name)

	err := a.istioAuthPolicyInterface(meta).Delete(meta.Name, &k8sMeta.DeleteOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return commons.HandleError(err, "error while deleting authentication policy: authentication policy not found")
		}
		return commons.HandleError(err, "error while deleting authentication policy")
	}

	log.Infof("Authentication policy deleted: %+v", meta.Name)
	return nil
}

func (a *istioImpl) istioAuthPolicyInterface(metaDto meta.Dto) istioAuthTyped.PolicyInterface {
	return a.istioAuthInterface.AuthenticationV1alpha1().Policies(metaDto.Namespace)
}

func (a *istioImpl) isEqual(oldRule *istioAuthApi.Policy, newRule *istioAuthApi.Policy) bool {
	return reflect.DeepEqual(oldRule.Spec, newRule.Spec)
}

func toIstioAuthPolicy(dto *Dto, defaultConfig JwtDefaultConfig, enableIstioAuthPolicyMtls bool) *istioAuthApi.Policy {

	objectMetadata := k8sMeta.ObjectMeta{
		Name:            dto.MetaDto.Name,
		Namespace:       dto.MetaDto.Namespace,
		Labels:          dto.MetaDto.Labels,
		UID:             dto.Status.Resource.Uid,
		ResourceVersion: dto.Status.Resource.Version,
	}

	var optionalPeers []*istioAuthApi.Peer = nil

	//For backward compatibility controlled by global enableIstioAuthPolicyMtls flag and api-specific DisablePolicyPeersMTLS override.
	if enableIstioAuthPolicyMtls && !dto.DisablePolicyPeersMTLS {
		optionalPeers = []*istioAuthApi.Peer{&istioAuthApi.Peer{}}
	}

	spec := &istioAuthApi.PolicySpec{
		Targets: []*istioAuthApi.Target{
			{Name: dto.ServiceName},
		},
		Peers: optionalPeers,
	}

	origins := make(istioAuthApi.Origins, 0, 1)

	if len(dto.Rules) != 0 {
		for _, rule := range dto.Rules {

			if rule.Type == JwtType {
				origins = append(origins, &istioAuthApi.Origin{
					Jwt: &istioAuthApi.Jwt{
						Issuer:  rule.Jwt.Issuer,
						JwksUri: rule.Jwt.JwksUri,
					},
				})
			}
		}
	} else if dto.AuthenticationEnabled {
		origins = append(origins, &istioAuthApi.Origin{
			Jwt: &istioAuthApi.Jwt{
				Issuer:  defaultConfig.Issuer,
				JwksUri: defaultConfig.JwksUri,
			},
		})
	}

	spec.Origins = origins

	spec.PrincipalBinding = istioAuthApi.UseOrigin

	return &istioAuthApi.Policy{
		ObjectMeta: objectMetadata,
		Spec:       spec,
	}
}

func isAuthenticationDisabled(dto *Dto) bool {
	return !dto.AuthenticationEnabled
}

func gatewayResourceFrom(policy *istioAuthApi.Policy) *kymaMeta.GatewayResource {
	return &kymaMeta.GatewayResource{
		Name:    policy.Name,
		Uid:     policy.UID,
		Version: policy.ResourceVersion,
	}
}
