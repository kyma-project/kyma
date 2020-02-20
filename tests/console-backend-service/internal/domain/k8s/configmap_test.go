// +build acceptance

package k8s

import (
	"testing"
	"time"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/retrier"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/waiter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	configMapName = "test-config-map"
)

type ConfigMapEvent struct {
	Type      string
	ConfigMap configMap
}

type configMapQueryResponse struct {
	ConfigMap configMap `json:"configMap"`
}

type configMapsQueryResponse struct {
	ConfigMaps []configMap `json:"configMaps"`
}

type updateConfigMapMutationResponse struct {
	UpdateConfigMap configMap `json:"updateConfigMap"`
}

type deleteConfigMapMutationResponse struct {
	DeleteConfigMap configMap `json:"deleteConfigMap"`
}

type configMap struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	CreationTimestamp int64  `json:"creationTimestamp"`
	Labels            labels `json:"labels"`
	JSON              json   `json:"json"`
}

func TestConfigMap(t *testing.T) {
	t.Skip("skipping unstable test")

	c, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Log("Subscribing to config maps...")
	subscription := c.Subscribe(fixConfigMapSubscription())
	defer subscription.Close()

	t.Log("Creating config map...")
	_, err = k8sClient.ConfigMaps(testNamespace).Create(fixConfigMap(configMapName))
	require.NoError(t, err)

	t.Log("Retrieving config map...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.ConfigMaps(testNamespace).Get(configMapName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Checking subscription for created config map...")
	expectedEvent := configMapEvent("ADD", configMap{Name: configMapName})
	assert.NoError(t, checkConfigMapEvent(expectedEvent, subscription))

	t.Log("Querying for config map...")
	var configMapRes configMapQueryResponse
	err = c.Do(fixConfigMapQuery(), &configMapRes)
	require.NoError(t, err)
	assert.Equal(t, configMapName, configMapRes.ConfigMap.Name)
	assert.Equal(t, testNamespace, configMapRes.ConfigMap.Namespace)

	t.Log("Querying for config maps...")
	var configMapsRes configMapsQueryResponse
	err = c.Do(fixConfigMapsQuery(), &configMapsRes)
	require.NoError(t, err)
	assert.Equal(t, configMapName, configMapsRes.ConfigMaps[0].Name)
	assert.Equal(t, testNamespace, configMapsRes.ConfigMaps[0].Namespace)

	t.Log("Updating...")
	var updateRes updateConfigMapMutationResponse
	err = retrier.Retry(func() error {
		err = c.Do(fixConfigMapQuery(), &configMapRes)
		if err != nil {
			return err
		}
		configMapRes.ConfigMap.JSON["metadata"].(map[string]interface{})["labels"] = map[string]string{"foo": "bar"}
		update, err := stringifyJSON(configMapRes.ConfigMap.JSON)
		if err != nil {
			return err
		}
		err = c.Do(fixUpdateConfigMapMutation(update), &updateRes)
		if err != nil {
			return err
		}
		return nil
	}, retrier.UpdateRetries)
	require.NoError(t, err)
	assert.Equal(t, configMapName, updateRes.UpdateConfigMap.Name)
	assert.Equal(t, testNamespace, updateRes.UpdateConfigMap.Namespace)

	t.Log("Checking subscription for updated config map...")
	expectedEvent = configMapEvent("UPDATE", configMap{Name: configMapName})
	assert.NoError(t, checkConfigMapEvent(expectedEvent, subscription))

	t.Log("Deleting config map...")
	var deleteRes deleteConfigMapMutationResponse
	err = c.Do(fixDeleteConfigMapMutation(), &deleteRes)
	require.NoError(t, err)
	assert.Equal(t, configMapName, deleteRes.DeleteConfigMap.Name)
	assert.Equal(t, testNamespace, deleteRes.DeleteConfigMap.Namespace)

	t.Log("Waiting for deletion...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.ConfigMaps(testNamespace).Get(configMapName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Checking subscription for deleted config map...")
	expectedEvent = configMapEvent("DELETE", configMap{Name: configMapName})
	assert.NoError(t, checkConfigMapEvent(expectedEvent, subscription))

	t.Log("Checking authorization directives...")
	as := auth.New()
	ops := &auth.OperationsInput{
		auth.Get:    {fixConfigMapQuery()},
		auth.List:   {fixConfigMapsQuery()},
		auth.Update: {fixUpdateConfigMapMutation("{\"\":\"\"}")},
		auth.Delete: {fixDeleteConfigMapMutation()},
	}
	as.Run(t, ops)
}

func fixConfigMap(name string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
	}
}

func fixConfigMapQuery() *graphql.Request {
	query := `query ($name: String!, $namespace: String!) {
				configMap(name: $name, namespace: $namespace) {
					name
					namespace
					creationTimestamp
					labels
					json
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("name", configMapName)
	req.SetVar("namespace", testNamespace)

	return req
}

func fixConfigMapsQuery() *graphql.Request {
	query := `query ($namespace: String!) {
				configMaps(namespace: $namespace) {
					name
					namespace
					creationTimestamp
					labels
					json
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", testNamespace)

	return req
}

func fixConfigMapSubscription() *graphql.Request {
	query := `subscription ($namespace: String!) {
				configMapEvent(namespace: $namespace) {
					type
					configMap {
						name
					}
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", testNamespace)

	return req
}

func fixUpdateConfigMapMutation(configMap string) *graphql.Request {
	mutation := `mutation ($name: String!, $namespace: String!, $configMap: JSON!) {
					updateConfigMap(name: $name, namespace: $namespace, configMap: $configMap) {
						name
						namespace
						creationTimestamp
						labels
						json
					}
				}`
	req := graphql.NewRequest(mutation)
	req.SetVar("name", configMapName)
	req.SetVar("namespace", testNamespace)
	req.SetVar("configMap", configMap)

	return req
}

func fixDeleteConfigMapMutation() *graphql.Request {
	mutation := `mutation ($name: String!, $namespace: String!) {
					deleteConfigMap(name: $name, namespace: $namespace) {
						name
						namespace
						creationTimestamp
						labels
						json
					}
				}`
	req := graphql.NewRequest(mutation)
	req.SetVar("name", configMapName)
	req.SetVar("namespace", testNamespace)

	return req
}

func configMapEvent(eventType string, configMap configMap) ConfigMapEvent {
	return ConfigMapEvent{
		Type:      eventType,
		ConfigMap: configMap,
	}
}

func readConfigMapEvent(sub *graphql.Subscription) (ConfigMapEvent, error) {
	type Response struct {
		ConfigMapEvent ConfigMapEvent
	}
	var configMapEvent Response
	err := sub.Next(&configMapEvent, tester.DefaultSubscriptionTimeout*2)

	return configMapEvent.ConfigMapEvent, err
}

func checkConfigMapEvent(expected ConfigMapEvent, sub *graphql.Subscription) error {
	for {
		event, err := readConfigMapEvent(sub)
		if err != nil {
			return err
		}
		if expected.Type == event.Type && expected.ConfigMap.Name == event.ConfigMap.Name {
			return nil
		}
	}
}
