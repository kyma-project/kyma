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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

const (
	SpecAPIType    = "API"
	SpecEventsType = "Events"

	connectedAppLabel = "connected-app"

	defaultDescription = "Description not provided"
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
)

type K8sResourceChecker struct {
	applicationClient       v1alpha1.ApplicationInterface
	clusterAssetGroupClient dynamic.ResourceInterface
	httpClient              *http.Client
}

func NewK8sResourceChecker(appClient v1alpha1.ApplicationInterface, clusterAssetGroupClient dynamic.ResourceInterface) *K8sResourceChecker {
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

func (c *K8sResourceChecker) AssertResourcesForApp(t *testing.T, logger *testkit.Logger, application compass.Application) { // TODO: remove t or remove checker
	appCR, err := c.applicationClient.Get(application.Name, v1meta.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, application.Name, appCR.Name)
	expectedDescription := defaultDescription
	if application.Description != nil && *application.Description != "" {
		expectedDescription = *application.Description
	}
	assert.Equal(t, expectedDescription, appCR.Spec.Description)

	assertLabels(t, application, appCR)

	for _, pkg := range application.Packages.Data {
		c.assertAPIPackageResources(t, logger, pkg, appCR)
	}
}

func assertLabels(t *testing.T, application compass.Application, appCR *v1alpha1apps.Application) {
	require.NotNil(t, appCR.Spec.Labels)
	assert.Equal(t, application.Name, appCR.Spec.Labels[connectedAppLabel])

	for key, value := range application.Labels {
		expectedValue := ""
		switch value.(type) {
		case string:
			expectedValue = value.(string)
			break
		case []string:
			expectedValue = strings.Join(value.([]string), ",")
			break
		}

		require.Equal(t, expectedValue, appCR.Spec.Labels[key])
	}
}

func (c *K8sResourceChecker) AssertAppResourcesDeleted(t *testing.T, applicationName string) {
	err := testkit.WaitForFunction(applicationDeletionCheckInterval, applicationDeletionTimeout, func() bool {
		_, err := c.applicationClient.Get(applicationName, v1meta.GetOptions{})
		if err == nil {
			return false
		}

		return k8serrors.IsNotFound(err)
	})
	require.NoError(t, err, fmt.Sprintf("Application %s not deleted", applicationName))
}

func (c *K8sResourceChecker) AssertAPIPackageResources(t *testing.T, logger *testkit.Logger, apiPackage *graphql.PackageExt, applicationName string) {
	appCR, err := c.applicationClient.Get(applicationName, v1meta.GetOptions{})
	require.NoError(t, err)

	c.assertAPIPackageResources(t, logger, apiPackage, appCR)
}

func (c *K8sResourceChecker) assertAPIPackageResources(t *testing.T, logger *testkit.Logger, apiPackage *graphql.PackageExt, appCR *v1alpha1apps.Application) {
	log := logger.NewExtended(map[string]string{"APIPackageId": apiPackage.ID, "APIPackageName": apiPackage.Name})
	log.Log("Assert resources for API Package")

	appCRSvc, found := getAppCRService(appCR, apiPackage.ID)
	require.True(t, found, log.ContextMsg("API Package not found in Application CR"))

	assert.Equal(t, apiPackage.ID, appCRSvc.ID)
	//assert.Equal(c.t, compassAPI.Name, appCRSvc.Name) // TODO: it is normalized name, either use same code or asset.NotEmpty
	assert.Equal(t, apiPackage.Name, appCRSvc.DisplayName)

	expectedDescription := defaultDescription
	if apiPackage.Description != nil && *apiPackage.Description != "" {
		expectedDescription = *apiPackage.Description
	}
	assert.Equal(t, expectedDescription, appCRSvc.Description)

	for _, api := range apiPackage.APIDefinitions.Data {
		c.assertAPI(t, log, *api, appCRSvc)
	}

	for _, eventAPI := range apiPackage.EventDefinitions.Data {
		c.assertEventAPI(t, log, *eventAPI, appCRSvc)
	}

	c.assertAssetGroup(t, apiPackage)
}

func (c *K8sResourceChecker) AssertAPIPackageDeleted(t *testing.T, logger *testkit.Logger, apiPackage *graphql.PackageExt, applicationName string) {
	log := logger.NewExtended(map[string]string{"APIPackageId": apiPackage.ID, "APIPackageName": apiPackage.Name})
	log.Log("Assert resources removed for API Package")

	appCR, err := c.applicationClient.Get(applicationName, v1meta.GetOptions{})
	require.NoError(t, err)

	_, found := getAppCRService(appCR, apiPackage.ID)
	assert.False(t, found, log.ContextMsg("API Package not removed from Application CR"))

	c.assertAssetGroupDeleted(t, apiPackage.ID)
}

func (c *K8sResourceChecker) assertAPI(t *testing.T, logger *testkit.Logger, compassAPI graphql.APIDefinitionExt, appCRSvc *v1alpha1apps.Service) {
	log := logger.NewExtended(map[string]string{"APIId": compassAPI.ID, "APIName": compassAPI.Name})
	log.Log("Assert resources for API")

	appCRSvcEntry, found := getAppCRServiceEntry(appCRSvc, compassAPI.ID)
	require.True(t, found, log.ContextMsg("API not found in Application CR Service"))

	assert.Equal(t, compassAPI.ID, appCRSvcEntry.ID)
	assert.Equal(t, compassAPI.Name, appCRSvcEntry.Name)
	assert.Equal(t, SpecAPIType, appCRSvcEntry.Type)
	assert.Equal(t, compassAPI.TargetURL, appCRSvcEntry.TargetUrl)
}

func (c *K8sResourceChecker) assertEventAPI(t *testing.T, logger *testkit.Logger, compassAPI graphql.EventAPIDefinitionExt, appCRSvc *v1alpha1apps.Service) {
	log := logger.NewExtended(map[string]string{"APIId": compassAPI.ID, "APIName": compassAPI.Name})
	log.Log("Assert resources for API")

	appCRSvcEntry, found := getAppCRServiceEntry(appCRSvc, compassAPI.ID)
	require.True(t, found, log.ContextMsg("API not found in Application CR Service"))

	assert.Equal(t, compassAPI.ID, appCRSvcEntry.ID)
	assert.Equal(t, compassAPI.Name, appCRSvcEntry.Name)
	assert.Equal(t, SpecEventsType, appCRSvcEntry.Type)
}

func (c *K8sResourceChecker) assertAssetGroupDeleted(t *testing.T, apiPackageId string) {
	_, err := c.clusterAssetGroupClient.Get(apiPackageId, v1meta.GetOptions{})
	require.Error(t, err)
	assert.True(t, k8serrors.IsNotFound(err))
}

func (c *K8sResourceChecker) assertAssetGroup(t *testing.T, apiPackage *graphql.PackageExt) {
	assetGroup := getClusterAssetGroup(t, apiPackage.ID, c.clusterAssetGroupClient)
	require.NotEmpty(t, assetGroup)

	for _, api := range apiPackage.APIDefinitions.Data {
		if apiSpecProvided(*api) {
			source, found := getAPISource(api.ID, assetGroup)
			require.True(t, found)

			c.assertAssetGroupSource(t, source, api.Spec.Format, string(*api.Spec.Data))
		}
	}

	for _, eventAPI := range apiPackage.EventDefinitions.Data {
		if eventAPISpecProvided(*eventAPI) {
			source, found := getAPISource(eventAPI.ID, assetGroup)
			require.True(t, found)

			c.assertAssetGroupSource(t, source, eventAPI.Spec.Format, string(*eventAPI.Spec.Data))
		}
	}

}

func apiSpecProvided(api graphql.APIDefinitionExt) bool {
	return api.Spec != nil && api.Spec.Data != nil && string(*api.Spec.Data) != ""
}

func eventAPISpecProvided(api graphql.EventAPIDefinitionExt) bool {
	return api.Spec != nil && api.Spec.Data != nil && string(*api.Spec.Data) != ""
}

func (c *K8sResourceChecker) assertAssetGroupSource(t *testing.T, source rafterapi.Source, format graphql.SpecFormat, expectedSpec string) {
	expectedSpecFormat := "." + strings.ToLower(format.String())
	assetGroupExtension := filepath.Ext(source.URL)
	assert.Equal(t, expectedSpecFormat, assetGroupExtension)

	c.checkContent(t, source.URL, expectedSpec)
}

func (c *K8sResourceChecker) checkContent(t *testing.T, url, expectedSpec string) {
	resp, err := c.httpClient.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, expectedSpec, string(bytes))
}

func getAPISource(apiID string, assetGroup rafterapi.ClusterAssetGroup) (rafterapi.Source, bool) {
	for _, source := range assetGroup.Spec.Sources {
		if strings.HasSuffix(string(source.Name), apiID) {
			return source, true
		}
	}
	return rafterapi.Source{}, false
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
