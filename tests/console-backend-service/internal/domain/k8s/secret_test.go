// +build acceptance

package k8s

import (
	"testing"
	"time"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/dex"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/waiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	secretName      = "test-secret"
	secretNamespace = "console-backend-service-secret"
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

type secret struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Data         string    `json:"data"`
	Type         string    `json:"type"`
	CreationTime time.Time `json:"creationTime"`
	Labels       string    `json:"labels"`
	Annotations  string    `json:"annotations"`
}

type secretStatusType string

func TestSecret(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	c, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Log("Creating namespace...")
	_, err = k8sClient.Namespaces().Create(fixNamespace(secretNamespace))
	require.NoError(t, err)

	defer func() {
		t.Log("Deleting namespace...")
		err = k8sClient.Namespaces().Delete(secretNamespace, &metav1.DeleteOptions{})
		require.NoError(t, err)
	}()

	t.Log("Subscribing to secrets...")
	subscription := c.Subscribe(fixSecretSubscription())
	defer subscription.Close()

	t.Log("Retrieving secret...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Querying for secret...")
	var secretRes secretQueryResponse
	err = c.Do(fixSecretQuery(), &secretRes)
	require.NoError(t, err)
	assert.Equal(t, secretName, secretRes.Secret.Name)
	assert.Equal(t, secretNamespace, secretRes.Secret.Namespace)

	t.Log("Querying for secrets...")
	var secretsRes secretsQueryResponse
	err = c.Do(fixSecretsQuery(), &secretsRes)
	require.NoError(t, err)
	assert.Equal(t, secretName, secretsRes.Secrets[0].Name)
	assert.Equal(t, secretNamespace, secretsRes.Secrets[0].Namespace)
}

func fixSecret(name, namespace string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: "type",
	}
}

func fixSecretQuery() *graphql.Request {
	query := `query ($name: String!, $namespace: String!) {
				secret(name: $name, namespace: $namespace) {
					name
					nodeName
					namespace
					restartCount
					creationTimestamp
					labels
					status
					containerStates {
						state
						reason
						message
					}
					json
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("name", secretName)
	req.SetVar("namespace", secretNamespace)

	return req
}

func fixSecretsQuery() *graphql.Request {
	query := `query ($namespace: String!) {
				secrets(namespace: $namespace) {
					name
					nodeName
					namespace
					restartCount
					creationTimestamp
					labels
					status
					containerStates {
						state
						reason
						message
					}
					json
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", secretNamespace)

	return req
}

func fixSecretSubscription() *graphql.Request {
	query := `subscription ($namespace: String!) {
				secretEvent(namespace: $namespace) {
					type
					secret {
						name
					}
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", secretNamespace)

	return req
}

func secretEvent(eventType string, secret secret) SecretEvent {
	return SecretEvent{
		Type:   eventType,
		Secret: secret,
	}
}

func readSecretEvent(sub *graphql.Subscription) (SecretEvent, error) {
	type Response struct {
		SecretEvent SecretEvent
	}
	var secretEvent Response
	err := sub.Next(&secretEvent, tester.DefaultSubscriptionTimeout)

	return secretEvent.SecretEvent, err
}

func checkSecretEvent(expected SecretEvent, sub *graphql.Subscription) error {
	for {
		event, err := readSecretEvent(sub)
		if err != nil {
			return err
		}
		if expected.Type == event.Type && expected.Secret.Name == event.Secret.Name {
			return nil
		}
	}
}
