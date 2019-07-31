package assertions

import (
	"testing"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/stretchr/testify/require"

	//application "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	//istio "github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"

	v1core "k8s.io/api/core/v1"
)

const (
	requestParametersHeadersKey         = "headers"
	requestParametersQueryParametersKey = "queryParameters"
)

type Labels map[string]string

type ServiceData struct {
	ServiceId            string
	DisplayName          string
	ProviderDisplayName  string
	LongDescription      string
	HasAPI               bool
	TargetUrl            string
	OauthUrl             string
	GatewayUrl           string
	AccessLabel          string
	HasEvents            bool
	CSRFTokenEndpointURL string
}

const (
	expectedProtocol v1core.Protocol = v1core.ProtocolTCP
	expectedPort     int32           = 80
)

type K8sResourceChecker struct {
	serviceClient v1.ServiceInterface
}

func NewK8sResourceChecker(serviceClient v1.ServiceInterface) *K8sResourceChecker {
	return &K8sResourceChecker{
		serviceClient: serviceClient,
	}
}

func (c *K8sResourceChecker) AssertResourcesForApp(t *testing.T, application compass.Application) {
	// TODO - do all other stuff
	//for _, api := range application.APIs.Data {
	//	c.CheckAPIK8sResources(t, api)
	//}

}

func (c *K8sResourceChecker) AssertAppResourcesDeleted(t *testing.T, appId string) {

}

func (c *K8sResourceChecker) AssertAPIResourcesDeleted(t *testing.T, apiId string) {

}

func (c *K8sResourceChecker) AssertEventAPIResourcesDeleted(t *testing.T, eventAPIId string) {

}

// TODO - take APIDefinition/EventAPIDefinition/Document/Application?
func (c *K8sResourceChecker) CheckAPIK8sResources(t *testing.T, api *graphql.APIDefinition) {
	resourceName := "app-" + api.ApplicationID + "-" + api.ID // TODO: this probably wont be enaugh

	c.checkK8sService(t, resourceName, nil, 8081) // TODO: target port
}

func (c *K8sResourceChecker) checkK8sService(t *testing.T, name string, labels Labels, targetPort int) {
	service, err := c.serviceClient.Get(name, v1meta.GetOptions{})
	require.NoError(t, err)

	require.Equal(t, name, service.Name)

	servicePorts := service.Spec.Ports[0]
	require.Equal(t, expectedProtocol, servicePorts.Protocol)
	require.Equal(t, int32(expectedPort), servicePorts.Port)
	require.Equal(t, int32(targetPort), servicePorts.TargetPort.IntVal)

	//checkLabels(t, labels, service.Labels) // TODO: how do we handle labels from Compass?
}

//func CheckK8sOAuthSecret(t *testing.T, secret *v1core.Secret, name string, labels Labels, clientId, clientSecret string) {
//	require.Equal(t, name, secret.Name)
//
//	secretData := secret.Data
//	require.Equal(t, clientId, string(secretData["clientId"]))
//	require.Equal(t, clientSecret, string(secretData["clientSecret"]))
//
//	checkLabels(t, labels, secret.Labels)
//}
//
//func CheckK8sBasicAuthSecret(t *testing.T, secret *v1core.Secret, name string, labels Labels, username, password string) {
//	require.Equal(t, name, secret.Name)
//
//	secretData := secret.Data
//	require.Equal(t, username, string(secretData["username"]))
//	require.Equal(t, password, string(secretData["password"]))
//
//	checkLabels(t, labels, secret.Labels)
//}
