package assertions

import (
	"strings"
	"testing"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"

	v1alpha1apps "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/stretchr/testify/require"

	//istio "github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"

	v1core "k8s.io/api/core/v1"
)

const (
	requestParametersHeadersKey         = "headers"
	requestParametersQueryParametersKey = "queryParameters"

	SpecAPIType          = "API"
	SpecEventsType       = "Events"
	CredentialsOAuthType = "OAuth"
	CredentialsBasicType = "Basic"
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
	expectedProtocol   v1core.Protocol = v1core.ProtocolTCP
	expectedPort       int32           = 80
	expectedTargetPort int32           = 8080
)

type K8sResourceChecker struct {
	serviceClient     v1.ServiceInterface
	secretClient      v1.SecretInterface
	applicationClient v1alpha1.ApplicationInterface
	nameResolver      *applications.NameResolver
}

func NewK8sResourceChecker(serviceClient v1.ServiceInterface, secretClient v1.SecretInterface, appClient v1alpha1.ApplicationInterface, nameResolver *applications.NameResolver) *K8sResourceChecker {
	return &K8sResourceChecker{
		serviceClient:     serviceClient,
		secretClient:      secretClient,
		applicationClient: appClient,
		nameResolver:      nameResolver,
	}
}

func (c *K8sResourceChecker) AssertResourcesForApp(t *testing.T, application compass.Application) {
	appCR, err := c.applicationClient.Get(application.ID, v1meta.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, application.ID, appCR.Name)
	expectedDescription := ""
	if application.Description != nil {
		expectedDescription = *application.Description
	}
	assert.Equal(t, expectedDescription, appCR.Spec.Description)

	// TODO - assert labels after proper handling in agent

	for _, api := range application.APIs.Data {
		c.assertAPI(t, *api, appCR)
	}

	// TODO - assert event apis

	// TODO - assert docs
}

func (c *K8sResourceChecker) AssertAppResourcesDeleted(t *testing.T, appId string) {
	_, err := c.applicationClient.Get(appId, v1meta.GetOptions{})
	require.Error(t, err)
	assert.True(t, k8serrors.IsNotFound(err))

	// TODO - should check all apis and stuff? Probably yes
}

func (c *K8sResourceChecker) AssertAPIResources(t *testing.T, compassAPI graphql.APIDefinition) {
	appCR, err := c.applicationClient.Get(compassAPI.ApplicationID, v1meta.GetOptions{})
	require.NoError(t, err)

	c.assertAPI(t, compassAPI, appCR)
}

// TODO - should take ID or whole API?
func (c *K8sResourceChecker) AssertAPIResourcesDeleted(t *testing.T, applicationId, apiId string) {
	logrus.Info("Checking API resources deleted for: ", apiId)
	appCR, err := c.applicationClient.Get(applicationId, v1meta.GetOptions{})
	require.NoError(t, err)

	for _, s := range appCR.Spec.Services {
		assert.NotEqual(t, s.ID, apiId)
	}
}

func (c *K8sResourceChecker) AssertEventAPIResources(t *testing.T, eventAPIId string) {

}

func (c *K8sResourceChecker) AssertEventAPIResourcesDeleted(t *testing.T, eventAPIId string) {

}

// TODO - Assert docs?

func (c *K8sResourceChecker) assertAPI(t *testing.T, compassAPI graphql.APIDefinition, appCR *v1alpha1apps.Application) {
	svc, found := getService(appCR, compassAPI.ID)
	assert.True(t, found)

	assert.True(t, strings.HasPrefix(svc.Name, compassAPI.Name))

	expectedDescription := "Description not provided"
	if compassAPI.Description != nil {
		expectedDescription = *compassAPI.Description
	}
	assert.Equal(t, expectedDescription, svc.Description)

	require.Equal(t, 1, len(svc.Entries))
	entry := svc.Entries[0]
	assert.Equal(t, SpecAPIType, entry.Type)
	assert.Equal(t, compassAPI.TargetURL, entry.TargetUrl)

	expectedGatewayURL := c.nameResolver.GetGatewayUrl(compassAPI.ApplicationID, compassAPI.ID)
	assert.Equal(t, expectedGatewayURL, entry.GatewayUrl)

	expectedResourceName := c.nameResolver.GetResourceName(compassAPI.ApplicationID, compassAPI.ID)
	assert.Equal(t, expectedResourceName, entry.AccessLabel)

	c.assertK8sService(t, expectedResourceName)

	if compassAPI.DefaultAuth != nil {
		c.assertCredentials(t, expectedResourceName, compassAPI.DefaultAuth, entry)
	}

	// TODO - assert Istio resources
	// TODO - assert Docs
}

func (c *K8sResourceChecker) assertEventAPI(t *testing.T, compassAPI graphql.APIDefinition, appCR *v1alpha1apps.Application) {
}

func (c *K8sResourceChecker) assertCredentials(t *testing.T, secretName string, auth *graphql.Auth, service v1alpha1apps.Entry) {
	switch cred := auth.Credential.(type) {
	case *graphql.BasicCredentialData:
		c.assertK8sBasicAuthSecret(t, secretName, cred, service) // TODO
	case *graphql.OAuthCredentialData:
		c.assertK8sOAuthSecret(t, secretName, cred, service) // TODO
	default:
		t.Fatalf("Unkonw credentials type")
	}

	c.assertCSRF(t, auth.RequestAuth, service)
}

func (c *K8sResourceChecker) assertK8sService(t *testing.T, name string) {
	service, err := c.serviceClient.Get(name, v1meta.GetOptions{})
	require.NoError(t, err)

	require.Equal(t, name, service.Name)

	servicePorts := service.Spec.Ports[0]
	require.Equal(t, expectedProtocol, servicePorts.Protocol)
	require.Equal(t, expectedPort, servicePorts.Port)
	require.Equal(t, expectedTargetPort, servicePorts.TargetPort.IntVal)
}

func (c *K8sResourceChecker) assertK8sBasicAuthSecret(t *testing.T, name string, credentials *graphql.BasicCredentialData, service v1alpha1apps.Entry) {
	secret, err := c.secretClient.Get(name, v1meta.GetOptions{})
	require.NoError(t, err)

	require.Equal(t, name, secret.Name)
	assert.Equal(t, name, service.Credentials.SecretName)
	assert.Equal(t, CredentialsBasicType, service.Credentials.Type)

	require.Equal(t, credentials.Username, string(secret.Data["username"]))
	require.Equal(t, credentials.Password, string(secret.Data["password"]))
}

func (c *K8sResourceChecker) assertK8sOAuthSecret(t *testing.T, name string, credentials *graphql.OAuthCredentialData, service v1alpha1apps.Entry) {
	secret, err := c.secretClient.Get(name, v1meta.GetOptions{})
	require.NoError(t, err)

	require.Equal(t, name, secret.Name)
	assert.Equal(t, name, service.Credentials.SecretName)
	assert.Equal(t, CredentialsOAuthType, service.Credentials.Type)
	assert.Equal(t, credentials.URL, service.Credentials.AuthenticationUrl)

	require.Equal(t, credentials.ClientID, string(secret.Data["clientId"]))
	require.Equal(t, credentials.ClientSecret, string(secret.Data["clientSecret"]))
}

func (c *K8sResourceChecker) assertCSRF(t *testing.T, auth *graphql.CredentialRequestAuth, service v1alpha1apps.Entry) {
	// TODO - implement
}

func getService(applicationCR *v1alpha1apps.Application, apiId string) (*v1alpha1apps.Service, bool) {
	for _, service := range applicationCR.Spec.Services {
		if service.ID == apiId {
			return &service, true
		}
	}

	return nil, false
}
