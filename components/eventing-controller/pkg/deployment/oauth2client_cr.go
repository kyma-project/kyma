package deployment

import (
	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const BEBSecretNameSuffix = "-beb-oauth2"

// NewOAuth2Client creates a Hydra OAuth2Client CR object
func NewOAuth2Client() *hydrav1alpha1.OAuth2Client {
	labels := map[string]string{
		AppLabelKey:       PublisherName,
		instanceLabelKey:  instanceLabelValue,
		dashboardLabelKey: dashboardLabelValue,
	}
	oa2CR := &hydrav1alpha1.OAuth2Client{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ControllerName,
			Namespace: ControllerNamespace,
			Labels:    labels,
		},
		Spec: hydrav1alpha1.OAuth2ClientSpec{
			GrantTypes: []hydrav1alpha1.GrantType{"client_credentials"},
			Scope:      "read write beb uaa.resource",
			SecretName: ControllerName + BEBSecretNameSuffix,
		},
	}
	return oa2CR
}
