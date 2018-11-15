package testkit

import (
	"testing"

	istio "github.com/kyma-project/kyma/components/metadata-service/pkg/apis/istio/v1alpha2"
	remoteenv "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/require"
	v1core "k8s.io/api/core/v1"
)

type Labels map[string]string

type ServiceData struct {
	ServiceId           string
	DisplayName         string
	ProviderDisplayName string
	LongDescription     string
	HasAPI              bool
	TargetUrl           string
	OauthUrl            string
	GatewayUrl          string
	AccessLabel         string
	HasEvents           bool
}

func CheckK8sService(t *testing.T, service *v1core.Service, name string, labels Labels, protocol v1core.Protocol, port, targetPort int) {
	require.Equal(t, name, service.Name)

	servicePorts := service.Spec.Ports[0]
	require.Equal(t, protocol, servicePorts.Protocol)
	require.Equal(t, int32(port), servicePorts.Port)
	require.Equal(t, int32(targetPort), servicePorts.TargetPort.IntVal)

	checkLabels(t, labels, service.Labels)
}

func CheckK8sSecret(t *testing.T, secret *v1core.Secret, name string, labels Labels, clientId, clientSecret string) {
	require.Equal(t, name, secret.Name)

	secretData := secret.Data
	require.Equal(t, clientId, string(secretData["clientId"]))
	require.Equal(t, clientSecret, string(secretData["clientSecret"]))

	checkLabels(t, labels, secret.Labels)
}

func CheckK8sIstioDenier(t *testing.T, denier *istio.Denier, name string, labels Labels, code int, message string) {
	require.Equal(t, name, denier.Name)

	denierStatus := denier.Spec.Status
	require.Equal(t, int32(code), denierStatus.Code)
	require.Equal(t, message, denierStatus.Message)

	checkLabels(t, labels, denier.Labels)
}

func CheckK8sIstioRule(t *testing.T, rule *istio.Rule, name, namespace string, labels Labels) {
	require.Equal(t, name, rule.Name)

	expectedMatchExpression := makeMatchExpression(name, namespace)
	require.Equal(t, expectedMatchExpression, rule.Spec.Match)

	ruleAction := rule.Spec.Actions[0]
	require.Equal(t, name+".denier", ruleAction.Handler)
	require.Equal(t, name+".checknothing", ruleAction.Instances[0])

	checkLabels(t, labels, rule.Labels)
}

func CheckK8sChecknothing(t *testing.T, checknothing *istio.Checknothing, name string, labels Labels) {
	require.Equal(t, name, checknothing.Name)

	checkLabels(t, labels, checknothing.Labels)
}

func CheckK8sRemoteEnvironment(t *testing.T, re *remoteenv.RemoteEnvironment, name string, expectedServiceData ServiceData) {
	require.Equal(t, name, re.Name)

	reService := findServiceInRemoteEnv(re.Spec.Services, expectedServiceData.ServiceId)
	require.NotNil(t, reService)

	require.Equal(t, expectedServiceData.ServiceId, reService.ID)
	require.Equal(t, expectedServiceData.DisplayName, reService.DisplayName)
	require.Equal(t, expectedServiceData.ProviderDisplayName, reService.ProviderDisplayName)
	require.Equal(t, expectedServiceData.LongDescription, reService.LongDescription)

	if expectedServiceData.HasAPI {
		apiEntry := findEntryOfType(reService.Entries, "API")
		require.NotNil(t, apiEntry)

		require.Equal(t, expectedServiceData.TargetUrl, apiEntry.TargetUrl)
		require.Equal(t, expectedServiceData.OauthUrl, apiEntry.OauthUrl)
		require.Equal(t, expectedServiceData.GatewayUrl, apiEntry.GatewayUrl)
		require.Equal(t, expectedServiceData.AccessLabel, apiEntry.AccessLabel)
	}

	if expectedServiceData.HasEvents {
		eventsEntry := findEntryOfType(reService.Entries, "Events")
		require.NotNil(t, eventsEntry)
	}
}

func CheckK8sRemoteEnvironmentNotContainsService(t *testing.T, re *remoteenv.RemoteEnvironment, serviceId string) {
	reService := findServiceInRemoteEnv(re.Spec.Services, serviceId)
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

func findServiceInRemoteEnv(reServices []remoteenv.Service, searchedID string) *remoteenv.Service {
	for _, e := range reServices {
		if e.ID == searchedID {
			return &e
		}
	}
	return nil
}

func findEntryOfType(entries []remoteenv.Entry, typeName string) *remoteenv.Entry {
	for _, e := range entries {
		if e.Type == typeName {
			return &e
		}
	}
	return nil
}
