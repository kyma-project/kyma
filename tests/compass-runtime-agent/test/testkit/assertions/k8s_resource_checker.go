package assertions

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	assets "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	istioclient "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"

	v1alpha1apps "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/stretchr/testify/require"

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
	serviceClient          v1.ServiceInterface
	secretClient           v1.SecretInterface
	applicationClient      v1alpha1.ApplicationInterface
	nameResolver           *applications.NameResolver
	istioClient            istioclient.Interface
	clusterDocsTopicClient dynamic.ResourceInterface
	integrationNamespace   string
}

func NewK8sResourceChecker(serviceClient v1.ServiceInterface, secretClient v1.SecretInterface, appClient v1alpha1.ApplicationInterface, nameResolver *applications.NameResolver,
	istioClient istioclient.Interface, clusterDocsTopicClient dynamic.ResourceInterface, integrationNamespace string) *K8sResourceChecker {
	return &K8sResourceChecker{
		serviceClient:          serviceClient,
		secretClient:           secretClient,
		applicationClient:      appClient,
		nameResolver:           nameResolver,
		istioClient:            istioClient,
		clusterDocsTopicClient: clusterDocsTopicClient,
		integrationNamespace:   integrationNamespace,
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

	// TODO: assert labels after proper handling in agent

	if len(application.APIs.Data) > 0 {
		c.assertAPIs(t, application.ID, application.APIs.Data, appCR)
	}
	if len(application.EventAPIs.Data) > 0 {
		c.assertEventAPIs(t, application.ID, application.EventAPIs.Data, appCR)
	}

	// TODO: assert Document after implementation in Director and Agent
}

func (c *K8sResourceChecker) AssertAppResourcesDeleted(t *testing.T, appId string) {
	_, err := c.applicationClient.Get(appId, v1meta.GetOptions{})
	require.Error(t, err, fmt.Sprintf("Application %s not deleted", appId))
	assert.True(t, k8serrors.IsNotFound(err))
}

func (c *K8sResourceChecker) AssertAPIResources(t *testing.T, appId string, compassAPIs ...*graphql.APIDefinition) {
	appCR, err := c.applicationClient.Get(appId, v1meta.GetOptions{})
	require.NoError(t, err)

	c.assertAPIs(t, appId, compassAPIs, appCR)
}

func (c *K8sResourceChecker) AssertAPIResourcesDeleted(t *testing.T, applicationId, apiId string) {
	appCR, err := c.applicationClient.Get(applicationId, v1meta.GetOptions{})
	require.NoError(t, err)

	for _, s := range appCR.Spec.Services {
		assert.NotEqual(t, s.ID, apiId)
	}

	c.assertServiceDeleted(t, applicationId, apiId)
}

func (c *K8sResourceChecker) AssertEventAPIResources(t *testing.T, appId string, compassEventAPIs ...*graphql.EventAPIDefinition) {
	appCR, err := c.applicationClient.Get(appId, v1meta.GetOptions{})
	require.NoError(t, err)

	c.assertEventAPIs(t, appId, compassEventAPIs, appCR)
}

func (c *K8sResourceChecker) assertAPIs(t *testing.T, appId string, compassAPIs []*graphql.APIDefinition, appCR *v1alpha1apps.Application) {
	for _, api := range compassAPIs {
		c.assertAPI(t, appId, *api, appCR)
	}
}

func (c *K8sResourceChecker) assertAPI(t *testing.T, appId string, compassAPI graphql.APIDefinition, appCR *v1alpha1apps.Application) {
	t.Logf("Asserting resources for %s API", compassAPI.ID)

	svc := c.assertService(t, compassAPI.ID, compassAPI.Name, compassAPI.Description, appCR)

	entry := svc.Entries[0]
	assert.Equal(t, SpecAPIType, entry.Type)
	assert.Equal(t, compassAPI.TargetURL, entry.TargetUrl)

	expectedGatewayURL := c.nameResolver.GetGatewayUrl(appId, compassAPI.ID)
	assert.Equal(t, expectedGatewayURL, entry.GatewayUrl)

	expectedResourceName := c.nameResolver.GetResourceName(appId, compassAPI.ID)
	assert.Equal(t, expectedResourceName, entry.AccessLabel)

	c.assertK8sService(t, expectedResourceName)

	if compassAPI.DefaultAuth != nil {
		c.assertCredentials(t, expectedResourceName, compassAPI.DefaultAuth, entry)
	}

	c.assertIstioResources(t, expectedResourceName, c.integrationNamespace)

	if apiSpecProvided(compassAPI) {
		c.assertDocsTopics(t, compassAPI.ID, string(*compassAPI.Spec.Data))
	}
}

func apiSpecProvided(api graphql.APIDefinition) bool {
	return api.Spec != nil && api.Spec.Data != nil && string(*api.Spec.Data) != ""
}

func (c *K8sResourceChecker) assertServiceDeleted(t *testing.T, applicationId, apiId string) {
	resourceName := c.nameResolver.GetResourceName(applicationId, apiId)
	c.assertResourcesDoNotExist(t, resourceName, apiId)
}

func (c *K8sResourceChecker) assertEventAPIs(t *testing.T, appId string, compassEventAPIs []*graphql.EventAPIDefinition, appCR *v1alpha1apps.Application) {
	for _, eventAPI := range compassEventAPIs {
		c.assertEventAPI(t, appId, *eventAPI, appCR)
	}
}

func (c *K8sResourceChecker) assertEventAPI(t *testing.T, appId string, compassEventAPI graphql.EventAPIDefinition, appCR *v1alpha1apps.Application) {
	svc := c.assertService(t, compassEventAPI.ID, compassEventAPI.Name, compassEventAPI.Description, appCR)

	entry := svc.Entries[0]
	assert.Equal(t, SpecEventsType, entry.Type)

	expectedResourceName := c.nameResolver.GetResourceName(appId, compassEventAPI.ID)
	assert.Equal(t, expectedResourceName, entry.AccessLabel)

	if eventAPISpecProvided(compassEventAPI) {
		c.assertDocsTopics(t, compassEventAPI.ID, string(*compassEventAPI.Spec.Data))
	}
}

func eventAPISpecProvided(api graphql.EventAPIDefinition) bool {
	return api.Spec != nil && api.Spec.Data != nil && string(*api.Spec.Data) != ""
}

func (c *K8sResourceChecker) assertService(t *testing.T, id, name string, description *string, appCR *v1alpha1apps.Application) *v1alpha1apps.Service {
	svc, found := getService(appCR, id)
	require.True(t, found)

	assert.True(t, strings.HasPrefix(svc.Name, name))

	expectedDescription := "Description not provided"
	if description != nil && *description != "" {
		expectedDescription = *description
	}
	assert.Equal(t, expectedDescription, svc.Description)

	require.Equal(t, 1, len(svc.Entries))

	return svc
}

func (c *K8sResourceChecker) assertCredentials(t *testing.T, secretName string, auth *graphql.Auth, service v1alpha1apps.Entry) {
	switch cred := auth.Credential.(type) {
	case *graphql.BasicCredentialData:
		c.assertK8sBasicAuthSecret(t, secretName, cred, service)
	case *graphql.OAuthCredentialData:
		c.assertK8sOAuthSecret(t, secretName, cred, service)
	default:
		t.Fatalf("Unknow credentials type")
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

func (c *K8sResourceChecker) assertResourcesDoNotExist(t *testing.T, resourceName, apiId string) {
	_, err := c.serviceClient.Get(resourceName, v1meta.GetOptions{})
	assert.Error(t, err)
	assert.True(t, k8serrors.IsNotFound(err))

	_, err = c.secretClient.Get(resourceName, v1meta.GetOptions{})
	assert.Error(t, err)
	assert.True(t, k8serrors.IsNotFound(err))

	//assert Istio resources have been removed
	_, err = c.istioClient.IstioV1alpha2().Rules(c.integrationNamespace).Get(resourceName, v1meta.GetOptions{})
	assert.Error(t, err)
	assert.True(t, k8serrors.IsNotFound(err))

	_, err = c.istioClient.IstioV1alpha2().Instances(c.integrationNamespace).Get(resourceName, v1meta.GetOptions{})
	assert.Error(t, err)
	assert.True(t, k8serrors.IsNotFound(err))

	_, err = c.istioClient.IstioV1alpha2().Handlers(c.integrationNamespace).Get(resourceName, v1meta.GetOptions{})
	assert.Error(t, err)
	assert.True(t, k8serrors.IsNotFound(err))

	//assert Docs Topics have been removed
	_, err = c.clusterDocsTopicClient.Get(apiId, v1meta.GetOptions{})
	assert.Error(t, err)
	assert.True(t, k8serrors.IsNotFound(err))
}

func (c *K8sResourceChecker) assertIstioResources(t *testing.T, resourceName string, namespace string) {
	alpha2 := c.istioClient.IstioV1alpha2()

	handler, e := alpha2.Handlers(namespace).Get(resourceName, v1meta.GetOptions{})
	require.NoError(t, e)
	require.NotEmpty(t, handler)

	instance, e := alpha2.Instances(namespace).Get(resourceName, v1meta.GetOptions{})
	require.NoError(t, e)
	require.NotEmpty(t, instance)

	rule, e := alpha2.Rules(namespace).Get(resourceName, v1meta.GetOptions{})
	require.NoError(t, e)
	require.NotEmpty(t, rule)
}

func (c *K8sResourceChecker) assertDocsTopics(t *testing.T, serviceID, expectedSpec string) {
	topic := getClusterDocsTopic(t, serviceID, c.clusterDocsTopicClient)
	require.NotEmpty(t, topic)
	require.NotEmpty(t, topic.Spec.Sources)
	checkContent(t, topic, expectedSpec)
}

func getService(applicationCR *v1alpha1apps.Application, apiId string) (*v1alpha1apps.Service, bool) {
	for _, service := range applicationCR.Spec.Services {
		if service.ID == apiId {
			return &service, true
		}
	}

	return nil, false
}

func getClusterDocsTopic(t *testing.T, id string, resourceInterface dynamic.ResourceInterface) assets.ClusterDocsTopic {
	u, err := resourceInterface.Get(id, v1meta.GetOptions{})
	require.NoError(t, err)

	var docsTopic assets.ClusterDocsTopic
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &docsTopic)
	require.NoError(t, err)

	return docsTopic
}

func checkContent(t *testing.T, topic assets.ClusterDocsTopic, expectedSpec string) {
	url := topic.Spec.Sources[0].URL
	resp, err := http.Get(url)
	defer resp.Body.Close()
	require.NoError(t, err)

	bytes, e := ioutil.ReadAll(resp.Body)
	require.NoError(t, e)

	assert.Equal(t, expectedSpec, string(bytes))
}
