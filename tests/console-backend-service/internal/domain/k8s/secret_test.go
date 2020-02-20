// +build acceptance

package k8s

import (
	"testing"
	"time"

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
	secretName = "test-secret"
	secretType = "test-secret-type"
)

type SecretEvent struct {
	Type   string
	Secret secret
}

type secretQueryResponse struct {
	Secret secret `json:"secret"`
}

type secretsQueryResponse struct {
	Secrets []secret `json:"secrets"`
}

type updateSecretMutationResponse struct {
	UpdateSecret secret `json:"updateSecret"`
}

type deleteSecretMutationResponse struct {
	DeleteSecret secret `json:"deleteSecret"`
}

type secret struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Data         json   `json:"data"`
	Type         string `json:"type"`
	CreationTime int    `json:"creationTime"`
	Labels       labels `json:"labels"`
	Annotations  json   `json:"annotations"`
	JSON         json   `json:"json"`
}

// TODO: Uncomment subscription tests once https://github.com/kyma-project/kyma/issues/3412 gets resolved

func TestSecret(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	//t.Log("Subscribing to secrets...")
	//subscription := c.Subscribe(fixSecretSubscription())
	//defer subscription.Close()

	t.Log("Creating secret...")
	_, err = k8sClient.Secrets(testNamespace).Create(fixSecret(secretName, testNamespace))
	require.NoError(t, err)

	t.Log("Retrieving secret...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Secrets(testNamespace).Get(secretName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)
	require.NoError(t, err)

	//t.Log("Checking subscription for created secret...")
	//expectedEvent := secretEvent("ADD", secret{Name: secretName})
	//assert.NoError(t, checkSecretEvent(expectedEvent, subscription))

	t.Log("Querying for secret...")
	var secretRes secretQueryResponse

	err = c.Do(fixSecretQuery(), &secretRes)
	require.NoError(t, err)
	assert.Equal(t, secretRes.Secret.Name, secretName)
	assert.Equal(t, secretRes.Secret.Namespace, testNamespace)

	t.Log("Querying for secrets...")
	var secretsRes secretsQueryResponse
	err = c.Do(fixSecretsQuery(), &secretsRes)
	require.NoError(t, err)

	var nameSlice []string
	var namespaceSlice []string

	for i := 0; i < len(secretsRes.Secrets); i++ {
		nameSlice = append(nameSlice, secretsRes.Secrets[i].Name)
		namespaceSlice = append(namespaceSlice, secretsRes.Secrets[i].Namespace)
	}

	assert.Contains(t, namespaceSlice, testNamespace)
	assert.Contains(t, nameSlice, secretName)

	t.Log("Updating...")
	var updateRes updateSecretMutationResponse
	err = retrier.Retry(func() error {
		err = c.Do(fixSecretQuery(), &secretRes)
		if err != nil {
			return err
		}
		secretRes.Secret.JSON["metadata"].(map[string]interface{})["labels"] = map[string]string{"foo": "bar"}
		update, err := stringifyJSON(secretRes.Secret.JSON)
		if err != nil {
			return err
		}
		err = c.Do(fixUpdateSecretMutation(update), &updateRes)
		if err != nil {
			return err
		}
		return nil
	}, retrier.UpdateRetries)
	require.NoError(t, err)
	assert.Equal(t, secretName, updateRes.UpdateSecret.Name)
	assert.Equal(t, testNamespace, updateRes.UpdateSecret.Namespace)

	//t.Log("Checking subscription for updated secret...")
	//expectedEvent = secretEvent("UPDATE", secret{Name: secretName})
	//assert.NoError(t, checkSecretEvent(expectedEvent, subscription))

	t.Log("Deleting secret...")
	var deleteRes deleteSecretMutationResponse
	err = c.Do(fixDeleteSecretMutation(), &deleteRes)
	require.NoError(t, err)
	assert.Equal(t, secretName, deleteRes.DeleteSecret.Name)
	assert.Equal(t, testNamespace, deleteRes.DeleteSecret.Namespace)

	t.Log("Waiting for deletion...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Secrets(testNamespace).Get(secretName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, time.Minute)
	require.NoError(t, err)

	//t.Log("Checking subscription for deleted secret...")
	//expectedEvent = secretEvent("DELETE", secret{Name: secretName})
	//assert.NoError(t, checkSecretEvent(expectedEvent, subscription))

	t.Log("Checking authorization directives...")
	as := auth.New()
	ops := &auth.OperationsInput{
		auth.Get:    {fixSecretQuery()},
		auth.List:   {fixSecretsQuery()},
		auth.Update: {fixUpdateSecretMutation("{\"\":\"\"}")},
		auth.Delete: {fixDeleteSecretMutation()},
	}
	as.Run(t, ops)
}

func fixSecret(name, namespace string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{"hey": "data"},
		},
		Type: secretType,
	}
}

func fixSecretQuery() *graphql.Request {
	query := `query ($name: String!, $namespace: String!) {
				secret(name: $name, namespace: $namespace) {
					name
					namespace
					creationTime
					labels
					data
					type
					annotations
					json
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("name", secretName)
	req.SetVar("namespace", testNamespace)

	return req
}

func fixSecretsQuery() *graphql.Request {
	query := `query ($namespace: String!) {
				secrets(namespace: $namespace) {
					name
					namespace
					creationTime
					labels
					data
					type
					annotations
					json
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", testNamespace)

	return req
}

//func fixSecretSubscription() *graphql.Request {
//	query := `subscription ($namespace: String!) {
//				secretEvent(namespace: $namespace) {
//					type
//					secret {
//						name
//					}
//				}
//			}`
//	req := graphql.NewRequest(query)
//	req.SetVar("namespace", testNamespace)
//
//	return req
//}

func fixUpdateSecretMutation(secret string) *graphql.Request {
	mutation := `mutation ($name: String!, $namespace: String!, $secret: JSON!) {
					updateSecret(name: $name, namespace: $namespace, secret: $secret) {
						name
						namespace
						creationTime
						labels
						data
						type
						annotations
						json
					}
				}`
	req := graphql.NewRequest(mutation)
	req.SetVar("name", secretName)
	req.SetVar("namespace", testNamespace)
	req.SetVar("secret", secret)

	return req
}

func fixDeleteSecretMutation() *graphql.Request {
	mutation := `mutation ($name: String!, $namespace: String!) {
					deleteSecret(name: $name, namespace: $namespace) {
						name
						namespace
						creationTime
						labels
						data
						type
						annotations
						json
					}
				}`
	req := graphql.NewRequest(mutation)
	req.SetVar("name", secretName)
	req.SetVar("namespace", testNamespace)

	return req
}

//func secretEvent(eventType string, secret secret) SecretEvent {
//	return SecretEvent{
//		Type:   eventType,
//		Secret: secret,
//	}
//}

//func readSecretEvent(sub *graphql.Subscription) (SecretEvent, error) {
//	type Response struct {
//		SecretEvent SecretEvent
//	}
//	var secretEvent Response
//	err := sub.Next(&secretEvent, tester.DefaultSubscriptionTimeout)
//
//	return secretEvent.SecretEvent, err
//}

//func checkSecretEvent(expected SecretEvent, sub *graphql.Subscription) error {
//	for {
//		event, err := readSecretEvent(sub)
//		if err != nil {
//			return err
//		}
//		if expected.Type == event.Type && expected.Secret.Name == event.Secret.Name {
//			return nil
//		}
//	}
//}
