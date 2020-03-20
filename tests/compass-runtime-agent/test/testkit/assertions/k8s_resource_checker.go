package assertions

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	v1alpha1apps "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	rafterapi "github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
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
	applicationDeletionTimeout       = 180 * time.Second
	applicationDeletionCheckInterval = 2 * time.Second

	expectedProtocol   v1core.Protocol = v1core.ProtocolTCP
	expectedPort       int32           = 80
	expectedTargetPort int32           = 8080
)

type K8sResourceChecker struct {
	applicationClient       v1alpha1.ApplicationInterface
	clusterAssetGroupClient dynamic.ResourceInterface
	httpClient              *http.Client
}

func NewK8sResourceChecker(appClient v1alpha1.ApplicationInterface, clusterAssetGroupClient dynamic.ResourceInterface, integrationNamespace string) *K8sResourceChecker {
	return &K8sResourceChecker{
		applicationClient:       appClient,
		clusterAssetGroupClient: clusterAssetGroupClient,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

type Checker struct {
	k8sChecker *K8sResourceChecker
	t          *testing.T
	log        *testkit.Logger
}

func (c *K8sResourceChecker) NewChecker(t *testing.T, log *testkit.Logger) *Checker {
	return &Checker{
		k8sChecker: c,
		t:          t,
		log:        log,
	}
}

func (c *Checker) AssertResourcesForApp(t *testing.T, application compass.Application) {
	appCR, err := c.k8sChecker.applicationClient.Get(application.Name, v1meta.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, application.Name, appCR.Name)
	expectedDescription := "Description not provided"
	if application.Description != nil && *application.Description != "" {
		expectedDescription = *application.Description
	}
	assert.Equal(t, expectedDescription, appCR.Spec.Description)

	for _, pkg := range application.Packages.Data {
		c.assertAPIPackageResources(pkg, appCR)
	}

	// TODO: assert labels after proper handling in agent

	// TODO: assert Document after implementation in Director and Agent
}

func (c *Checker) AssertAppResourcesDeleted(t *testing.T, applicationName string) {
	err := testkit.WaitForFunction(applicationDeletionCheckInterval, applicationDeletionTimeout, func() bool {
		_, err := c.k8sChecker.applicationClient.Get(applicationName, v1meta.GetOptions{})
		if err == nil {
			return false
		}

		return k8serrors.IsNotFound(err)
	})
	require.NoError(t, err, fmt.Sprintf("Application %s not deleted", applicationName))
}

//func (c *Checker) AssertAPIResources(t *testing.T, applicationName string, compassAPIs ...*graphql.APIDefinitionExt) {
//	appCR, err := c.k8sChecker.applicationClient.Get(applicationName, v1meta.GetOptions{})
//	require.NoError(t, err)
//
//	c.assertAPIs(t, applicationName, compassAPIs, appCR)
//}

//func (c *Checker) AssertAPIResourcesDeleted(t *testing.T, applicationName, apiId string) {
//	appCR, err := c.k8sChecker.applicationClient.Get(applicationName, v1meta.GetOptions{})
//	require.NoError(t, err)
//
//	for _, s := range appCR.Spec.Services {
//		assert.NotEqual(t, s.ID, apiId)
//	}
//
//	//c.assertServiceDeleted(t, applicationName, apiId)
//}

//func (c *Checker) AssertEventAPIResources(t *testing.T, applicationName string, compassEventAPIs ...*graphql.EventAPIDefinitionExt) {
//	appCR, err := c.k8sChecker.applicationClient.Get(applicationName, v1meta.GetOptions{})
//	require.NoError(t, err)
//
//	c.assertEventAPIs(t, applicationName, compassEventAPIs, appCR)
//}

func (c *Checker) AssertAPIPackageResources(apiPackage *graphql.PackageExt, applicationName string) {
	appCR, err := c.k8sChecker.applicationClient.Get(applicationName, v1meta.GetOptions{})
	require.NoError(c.t, err)

	c.assertAPIPackageResources(apiPackage, appCR)
}

func (c *Checker) assertAPIPackageResources(apiPackage *graphql.PackageExt, appCR *v1alpha1apps.Application) {
	log := c.log.NewExtended(map[string]string{"APIPackageId": apiPackage.ID, "APIPackageName": apiPackage.Name})
	log.Log("Assert resources for API Package")

	appCRSvc, found := getAppCRService(appCR, apiPackage.ID)
	require.True(c.t, found, log.ContextMsg("API Package not found in Application CR"))

	assert.Equal(c.t, apiPackage.ID, appCRSvc.ID)
	//assert.Equal(c.t, compassAPI.Name, appCRSvc.Name) // TODO: it is normalized name, either use same code or asset.NotEmpty
	assert.Equal(c.t, apiPackage.Name, appCRSvc.DisplayName)

	expectedDescription := ""
	if apiPackage.Description != nil {
		expectedDescription = *apiPackage.Description
	}
	assert.Equal(c.t, expectedDescription, appCRSvc.Description)

	for _, api := range apiPackage.APIDefinitions.Data {
		c.assertAPI(*api, appCRSvc)
	}

	for _, eventAPI := range apiPackage.EventDefinitions.Data {
		c.assertEventAPI(*eventAPI, appCRSvc)
	}

	// TODO: asset groups
}

func (c *Checker) AssertAPIPackageDeleted(apiPackage *graphql.PackageExt, applicationName string) {
	log := c.log.NewExtended(map[string]string{"APIPackageId": apiPackage.ID, "APIPackageName": apiPackage.Name})
	log.Log("Assert resources removed for API Package")

	appCR, err := c.k8sChecker.applicationClient.Get(applicationName, v1meta.GetOptions{})
	require.NoError(c.t, err)

	_, found := getAppCRService(appCR, apiPackage.ID)
	assert.False(c.t, found, log.ContextMsg("API Package not removed from Application CR"))
}

func (c *Checker) assertAPI(compassAPI graphql.APIDefinitionExt, appCRSvc *v1alpha1apps.Service) {
	log := c.log.NewExtended(map[string]string{"APIId": compassAPI.ID, "APIName": compassAPI.Name})
	log.Log("Assert resources for API")

	appCRSvcEntry, found := getAppCRServiceEntry(appCRSvc, compassAPI.ID)
	require.True(c.t, found, log.ContextMsg("API not found in Application CR Service"))

	assert.Equal(c.t, compassAPI.ID, appCRSvcEntry.ID)
	assert.Equal(c.t, compassAPI.Name, appCRSvcEntry.Name)
	assert.Equal(c.t, SpecAPIType, appCRSvcEntry.Type)
	assert.Equal(c.t, compassAPI.TargetURL, appCRSvcEntry.TargetUrl)

	//if apiSpecProvided(compassAPI) {
	//	assert.Equal(c.t, compassAPI.Spec.Type.String(), appCRSvcEntry.ApiType)
	//	c.assertAssetGroup(c.t, compassAPI.ID, string(*compassAPI.Spec.Data), string(compassAPI.Spec.Format))
	//}
}

func apiSpecProvided(api graphql.APIDefinitionExt) bool {
	return api.Spec != nil && api.Spec.Data != nil && string(*api.Spec.Data) != ""
}

//func (c *Checker) assertServiceDeleted(t *testing.T, applicationName, apiId string) {
//	resourceName := c.k8sChecker.nameResolver.GetResourceName(applicationName, apiId)
//	c.assertResourcesDoNotExist(t, resourceName, apiId)
//}

func (c *Checker) assertEventAPI(compassAPI graphql.EventAPIDefinitionExt, appCRSvc *v1alpha1apps.Service) {
	log := c.log.NewExtended(map[string]string{"APIId": compassAPI.ID, "APIName": compassAPI.Name})
	log.Log("Assert resources for API")

	appCRSvcEntry, found := getAppCRServiceEntry(appCRSvc, compassAPI.ID)
	require.True(c.t, found, log.ContextMsg("API not found in Application CR Service"))

	assert.Equal(c.t, compassAPI.ID, appCRSvcEntry.ID)
	assert.Equal(c.t, compassAPI.Name, appCRSvcEntry.Name)
	assert.Equal(c.t, SpecEventsType, appCRSvcEntry.Type)

	//if eventAPISpecProvided(compassAPI) {
	//	assert.Equal(c.t, compassAPI.Spec.Type.String(), appCRSvcEntry.ApiType)
	//	c.assertAssetGroup(c.t, compassAPI.ID, string(*compassAPI.Spec.Data), string(compassAPI.Spec.Format))
	//}
}

func eventAPISpecProvided(api graphql.EventAPIDefinitionExt) bool {
	return api.Spec != nil && api.Spec.Data != nil && string(*api.Spec.Data) != ""
}

//func (c *Checker) assertK8sService(t *testing.T, name string) {
//	service, err := c.k8sChecker.serviceClient.Get(name, v1meta.GetOptions{})
//	require.NoError(t, err)
//
//	require.Equal(t, name, service.Name)
//
//	servicePorts := service.Spec.Ports[0]
//	require.Equal(t, expectedProtocol, servicePorts.Protocol)
//	require.Equal(t, expectedPort, servicePorts.Port)
//	require.Equal(t, expectedTargetPort, servicePorts.TargetPort.IntVal)
//}

//func (c *Checker) assertK8sBasicAuthSecret(t *testing.T, name string, credentials *graphql.BasicCredentialData, service v1alpha1apps.Entry) {
//	secret, err := c.k8sChecker.secretClient.Get(name, v1meta.GetOptions{})
//	require.NoError(t, err)
//
//	require.Equal(t, name, secret.Name)
//	assert.Equal(t, name, service.Credentials.SecretName)
//	assert.Equal(t, CredentialsBasicType, service.Credentials.Type)
//
//	require.Equal(t, credentials.Username, string(secret.Data["username"]))
//	require.Equal(t, credentials.Password, string(secret.Data["password"]))
//}
//
//func (c *Checker) assertK8sOAuthSecret(t *testing.T, name string, credentials *graphql.OAuthCredentialData, service v1alpha1apps.Entry) {
//	secret, err := c.k8sChecker.secretClient.Get(name, v1meta.GetOptions{})
//	require.NoError(t, err)
//
//	require.Equal(t, name, secret.Name)
//	assert.Equal(t, name, service.Credentials.SecretName)
//	assert.Equal(t, CredentialsOAuthType, service.Credentials.Type)
//	assert.Equal(t, credentials.URL, service.Credentials.AuthenticationUrl)
//
//	require.Equal(t, credentials.ClientID, string(secret.Data["clientId"]))
//	require.Equal(t, credentials.ClientSecret, string(secret.Data["clientSecret"]))
//}

//func (c *Checker) assertResourcesDoNotExist(t *testing.T, resourceName, apiId string) {
//	_, err := c.k8sChecker.serviceClient.Get(resourceName, v1meta.GetOptions{})
//	assert.Error(t, err)
//	assert.True(t, k8serrors.IsNotFound(err))
//
//	_, err = c.k8sChecker.secretClient.Get(resourceName, v1meta.GetOptions{})
//	assert.Error(t, err)
//	assert.True(t, k8serrors.IsNotFound(err))
//
//	//assert Istio resources have been removed
//	_, err = c.k8sChecker.istioHandlersClient.Get(resourceName, v1meta.GetOptions{})
//	assert.Error(t, err)
//	assert.True(t, k8serrors.IsNotFound(err))
//
//	_, err = c.k8sChecker.istioInstancesClient.Get(resourceName, v1meta.GetOptions{})
//	assert.Error(t, err)
//	assert.True(t, k8serrors.IsNotFound(err))
//
//	_, err = c.k8sChecker.istioRulesClient.Get(resourceName, v1meta.GetOptions{})
//	assert.Error(t, err)
//	assert.True(t, k8serrors.IsNotFound(err))
//
//	//assert Asset Groups have been removed
//	_, err = c.k8sChecker.clusterAssetGroupClient.Get(apiId, v1meta.GetOptions{})
//	assert.Error(t, err)
//	assert.True(t, k8serrors.IsNotFound(err))
//}

func (c *Checker) assertAssetGroup(t *testing.T, serviceID, expectedSpec, specFormat string) {
	assetGroup := getClusterAssetGroup(t, serviceID, c.k8sChecker.clusterAssetGroupClient)
	assetGroupSourceURL := assetGroup.Spec.Sources[0].URL
	assetGroupExtension := filepath.Ext(assetGroupSourceURL)
	expectedSpecFormat := "." + strings.ToLower(specFormat)

	require.NotEmpty(t, assetGroup)
	require.NotEmpty(t, assetGroup.Spec.Sources)
	require.Equal(t, expectedSpecFormat, assetGroupExtension)
	c.checkContent(t, assetGroup, expectedSpec)
}

func getAppCRService(applicationCR *v1alpha1apps.Application, apiPkgId string) (*v1alpha1apps.Service, bool) {
	for _, service := range applicationCR.Spec.Services {
		if service.ID == apiPkgId {
			return &service, true
		}
	}

	return nil, false
}

func getAppCRServiceEntry(appCRSvc *v1alpha1apps.Service, apiId string) (*v1alpha1apps.Entry, bool) {
	for _, entry := range appCRSvc.Entries {
		if entry.ID == apiId {
			return &entry, true
		}
	}

	return nil, false
}

func getClusterAssetGroup(t *testing.T, id string, resourceInterface dynamic.ResourceInterface) rafterapi.ClusterAssetGroup {
	u, err := resourceInterface.Get(id, v1meta.GetOptions{})
	require.NoError(t, err)

	var assetGroup rafterapi.ClusterAssetGroup
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &assetGroup)
	require.NoError(t, err)

	return assetGroup
}

func (c *Checker) checkContent(t *testing.T, assetGroup rafterapi.ClusterAssetGroup, expectedSpec string) {
	url := assetGroup.Spec.Sources[0].URL

	resp, err := c.k8sChecker.httpClient.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	bytes, e := ioutil.ReadAll(resp.Body)
	require.NoError(t, e)

	assert.Equal(t, expectedSpec, string(bytes))
}
