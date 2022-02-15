package testkit

import (
	"crypto/tls"
	"encoding/json"
	"testing"

	application "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/require"
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

func CheckK8sService(t *testing.T, service *v1core.Service, name string, labels Labels, protocol v1core.Protocol, port, targetPort int) {
	require.Equal(t, name, service.Name)

	servicePorts := service.Spec.Ports[0]
	require.Equal(t, protocol, servicePorts.Protocol)
	require.Equal(t, int32(port), servicePorts.Port)
	require.Equal(t, int32(targetPort), servicePorts.TargetPort.IntVal)

	checkLabels(t, labels, service.Labels)
}

func CheckK8sOAuthSecret(t *testing.T, secret *v1core.Secret, name string, labels Labels, clientId, clientSecret string) {
	require.Equal(t, name, secret.Name)

	secretData := secret.Data
	require.Equal(t, clientId, string(secretData["clientId"]))
	require.Equal(t, clientSecret, string(secretData["clientSecret"]))

	checkLabels(t, labels, secret.Labels)
}

func CheckK8sBasicAuthSecret(t *testing.T, secret *v1core.Secret, name string, labels Labels, username, password string) {
	require.Equal(t, name, secret.Name)

	secretData := secret.Data
	require.Equal(t, username, string(secretData["username"]))
	require.Equal(t, password, string(secretData["password"]))

	checkLabels(t, labels, secret.Labels)
}

func CheckK8sCertificateGenSecret(t *testing.T, secret *v1core.Secret, name string, labels Labels, commonName string) {
	require.Equal(t, name, secret.Name)

	secretData := secret.Data
	require.Equal(t, commonName, string(secretData["commonName"]))

	crt := secretData["crt"]
	key := secretData["key"]

	cert, err := tls.X509KeyPair(crt, key)
	require.NoError(t, err)
	require.Nil(t, cert.Leaf)

	checkLabels(t, labels, secret.Labels)
}

func CheckK8sParamsSecret(t *testing.T, secret *v1core.Secret, name string, labels Labels, headerKey, headerValue, queryParameterKey, queryParameterValue string) {
	require.Equal(t, name, secret.Name)

	secretData := secret.Data
	var requestParameters RequestParameters

	rawHeaders := secretData[requestParametersHeadersKey]
	err := json.Unmarshal(rawHeaders, &requestParameters.Headers)
	require.NoError(t, err)
	headers := *requestParameters.Headers

	rawQueryParameters := secretData[requestParametersQueryParametersKey]
	err = json.Unmarshal(rawQueryParameters, &requestParameters.QueryParameters)
	require.NoError(t, err)
	queryParameters := *requestParameters.QueryParameters

	require.Equal(t, headerValue, headers[headerKey][0])
	require.Equal(t, queryParameterValue, queryParameters[queryParameterKey][0])

	checkLabels(t, labels, secret.Labels)
}

func CheckK8sApplication(t *testing.T, app *application.Application, name string, expectedServiceData ServiceData) {
	require.Equal(t, name, app.Name)

	appService := findServiceInApp(app.Spec.Services, expectedServiceData.ServiceId)
	require.NotNil(t, appService)

	require.Equal(t, expectedServiceData.ServiceId, appService.ID)
	require.Equal(t, expectedServiceData.DisplayName, appService.DisplayName)
	require.Equal(t, expectedServiceData.ProviderDisplayName, appService.ProviderDisplayName)
	require.Equal(t, expectedServiceData.LongDescription, appService.LongDescription)

	if expectedServiceData.HasAPI {
		apiEntry := findEntryOfType(appService.Entries, "API")
		require.NotNil(t, apiEntry)

		if apiEntry.Type == "OAuth" {
			require.Equal(t, expectedServiceData.OauthUrl, apiEntry.Credentials.AuthenticationUrl)
		}

		if apiEntry.Credentials.CSRFInfo != nil {
			require.Equal(t, expectedServiceData.CSRFTokenEndpointURL, apiEntry.Credentials.CSRFInfo.TokenEndpointURL)
		}

		require.Equal(t, expectedServiceData.TargetUrl, apiEntry.TargetUrl)
		require.Equal(t, expectedServiceData.GatewayUrl, apiEntry.GatewayUrl)
		require.Equal(t, expectedServiceData.AccessLabel, apiEntry.AccessLabel)
	}

	if expectedServiceData.HasEvents {
		eventsEntry := findEntryOfType(appService.Entries, "Events")
		require.NotNil(t, eventsEntry)
	}
}

func CheckK8sApplicationNotContainsService(t *testing.T, re *application.Application, serviceId string) {
	reService := findServiceInApp(re.Spec.Services, serviceId)
	require.Nil(t, reService)
}

func checkLabels(t *testing.T, expected, actual Labels) {
	for key := range expected {
		require.Equal(t, expected[key], actual[key])
	}
}

func makeMatchExpression(name, namespace string) string {
	return `(destination.service.host == "` + name + "." + namespace + `.svc.cluster.local") && (source.labels["` + name + `"] != "true")`
}

func findServiceInApp(reServices []application.Service, searchedID string) *application.Service {
	for _, e := range reServices {
		if e.ID == searchedID {
			return &e
		}
	}
	return nil
}

func findEntryOfType(entries []application.Entry, typeName string) *application.Entry {
	for _, e := range entries {
		if e.Type == typeName {
			return &e
		}
	}
	return nil
}
