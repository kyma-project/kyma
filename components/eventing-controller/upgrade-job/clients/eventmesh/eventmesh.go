package eventmesh

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	emsclient "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/auth"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/backend"
)

// Client struct for Event Mesh client
type Client struct {
	client *emsclient.Client
}

// NewClient creates and returns new client for EventMesh (BEB)
func NewClient() Client {
	return Client{}
}

// InitUsingSecret initializes the client by parsing the provided BEB secret (for v1.24.x)
func (c *Client) InitUsingSecret(secret *corev1.Secret) error {
	// First set beb config
	cfg, err := c.processSecret(secret)
	if err != nil {
		return err
	}

	// Setup authenticator and beb client
	if c.client == nil {
		authenticator := auth.NewAuthenticator(cfg)
		c.client = emsclient.NewClient(config.GetDefaultConfig(cfg.BebApiUrl), authenticator)
	}
	return nil
}

// IsInitialised returns true if the BEB client is initialised
// or returns false if not initialised
func (c *Client) IsInitialised() bool {
	if c.client != nil {
		return true
	}
	return false
}

// Delete deletes the specified subscription from Event Mesh (BEB).
// It returns DeleteResponse, nil if the request was successful.
// The DeleteResponse object contains the information if the deletion was successful or not
// or returns an error if it fails to send the request
func (c *Client) Delete(bebSubscriptionName string) (*types.DeleteResponse, error) {
	return c.client.Delete(bebSubscriptionName)
}

// processSecret private method to process BEB secret
func (c *Client) processSecret(bebSecret *corev1.Secret) (env.Config, error) {
	cfg := env.Config{}

	// First parse/decode the bebsecret
	secret, err := c.getParsedSecret(bebSecret)
	if err != nil {
		return cfg, err
	}

	// Read the config values
	cfg.BebApiUrl = fmt.Sprintf("%s%s", string(secret.StringData[backend.PublisherSecretEMSHostKey]), backend.BEBPublishEndpointForSubscriber)
	if len(cfg.BebApiUrl) == 0 {
		return cfg, errors.New("cannot get BebApiUrl from BEB secret")
	}

	cfg.ClientID = secret.StringData[deployment.PublisherSecretClientIDKey]
	if len(cfg.ClientID) == 0 {
		return cfg, errors.New("cannot get CLIENT_ID from BEB secret")
	}

	cfg.ClientSecret = string(secret.StringData[deployment.PublisherSecretClientSecretKey])
	if len(cfg.ClientSecret) == 0 {
		return cfg, errors.New("cannot get CLIENT_SECRET from BEB secret")
	}

	cfg.TokenEndpoint = string(secret.StringData[deployment.PublisherSecretTokenEndpointKey])
	if len(cfg.TokenEndpoint) == 0 {
		return cfg, errors.New("cannot get TOKEN_ENDPOINT from BEB secret")
	}

	cfg.BEBNamespace = fmt.Sprintf("%s%s", backend.NamespacePrefix, string(secret.StringData[deployment.PublisherSecretBEBNamespaceKey]))
	if len(cfg.BEBNamespace) == 0 {
		return cfg, errors.New("cannot get BEB_NAMESPACE from BEB secret")
	}

	return cfg, nil
}

// getParsedSecret private method to parse/decode BEB secret
func (c *Client) getParsedSecret(bebSecret *corev1.Secret) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bebSecret.Name,
			Namespace: bebSecret.Namespace,
		},
	}

	if _, ok := bebSecret.Data["messaging"]; !ok {
		return nil, errors.New("message is missing from BEB secret")
	}
	messagingBytes := bebSecret.Data["messaging"]

	if _, ok := bebSecret.Data["namespace"]; !ok {
		return nil, errors.New("namespace is missing from BEB secret")
	}
	namespaceBytes := bebSecret.Data["namespace"]

	var messages []backend.Message
	err := json.Unmarshal(messagingBytes, &messages)
	if err != nil {
		return nil, err
	}

	for _, m := range messages {
		if m.Broker.BrokerType == "saprestmgw" {
			if len(m.OA2.ClientID) == 0 {
				return nil, errors.New("client ID is missing")
			}
			if len(m.OA2.ClientSecret) == 0 {
				return nil, errors.New("client secret is missing")
			}
			if len(m.OA2.TokenEndpoint) == 0 {
				return nil, errors.New("tokenendpoint is missing")
			}
			if len(m.OA2.GrantType) == 0 {
				return nil, errors.New("granttype is missing")
			}
			if len(m.URI) == 0 {
				return nil, errors.New("publish URL is missing")
			}

			secret.StringData = c.getSecretStringData(m.OA2.ClientID, m.OA2.ClientSecret, m.OA2.TokenEndpoint, m.OA2.GrantType, m.URI, string(namespaceBytes))
			break
		}
	}

	return secret, nil
}

// getSecretStringData creates Map object for provided data
func (c *Client) getSecretStringData(clientID, clientSecret, tokenEndpoint, grantType, publishURL, namespace string) map[string]string {
	return map[string]string{
		deployment.PublisherSecretClientIDKey:      clientID,
		deployment.PublisherSecretClientSecretKey:  clientSecret,
		deployment.PublisherSecretTokenEndpointKey: fmt.Sprintf(backend.TokenEndpointFormat, tokenEndpoint, grantType),
		deployment.PublisherSecretEMSURLKey:        fmt.Sprintf("%s%s", publishURL, backend.BEBPublishEndpointForPublisher),
		backend.PublisherSecretEMSHostKey:          fmt.Sprintf("%s", publishURL),
		deployment.PublisherSecretBEBNamespaceKey:  namespace,
	}
}
