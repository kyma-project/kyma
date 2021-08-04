package eventmesh

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
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
	//authenticator := auth.NewAuthenticator(cfg)
	// client.NewClient(config.GetDefaultConfig(cfg.BebApiUrl), authenticator)
	return Client{}
}

// Init initializes the client by parsing the provided BEB secret
// It needs to be called once before the client can be used for further operations.
func (c *Client) Init(secret *v1.Secret) error {
	// First set beb config
	cfg := env.Config{}
	err := c.processSecret(&cfg, secret)
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

// Delete deletes the specified subcription from Event Mesh (BEB).
// It returns DeleteResponse, nil if the request was successful.
// The DeleteResponse object contains the information if the deletion was successful or not
// or returns an error if it fails to send the request
func (c *Client) Delete(bebSubscriptionName string) (*types.DeleteResponse, error) {
	return c.client.Delete(bebSubscriptionName)
}

// processSecret private method to process BEB secret
func (c *Client) processSecret(cfg *env.Config, bebSecret *v1.Secret) error {
	// First parse/decode the bebsecret
	secret, err := c.getParsedSecret(bebSecret)
	if err != nil {
		return err
	}

	// Read the config values
	cfg.BebApiUrl = fmt.Sprintf("%s%s", string(secret.StringData[backend.PublisherSecretEMSHostKey]), backend.BEBPublishEndpointForSubscriber)
	if len(cfg.BebApiUrl) == 0 {
		return errors.New("cannot get BEB_API_URL env var")
	}

	cfg.ClientID = string(secret.StringData[deployment.PublisherSecretClientIDKey])
	if len(cfg.ClientID) == 0 {
		return errors.New("cannot get CLIENT_ID env var")
	}

	cfg.ClientSecret = string(secret.StringData[deployment.PublisherSecretClientSecretKey])
	if len(cfg.ClientSecret) == 0 {
		return errors.New( "cannot get CLIENT_SECRET env var")
	}

	cfg.TokenEndpoint= string(secret.StringData[deployment.PublisherSecretTokenEndpointKey])
	if len(cfg.TokenEndpoint) == 0 {
		return errors.New("cannot get TOKEN_ENDPOINT env var")
	}

	cfg.BEBNamespace = fmt.Sprintf("%s%s", backend.NamespacePrefix, string(secret.StringData[deployment.PublisherSecretBEBNamespaceKey]))
	if len(cfg.BEBNamespace) == 0 {
		return errors.New( "cannot get BEB_NAMESPACE env var")
	}
	return nil
}

// getParsedSecret private method to parse/decode BEB secret
func (c *Client) getParsedSecret(bebSecret *v1.Secret) (*v1.Secret, error) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "temp",
			Namespace: "temp",
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
		backend.PublisherSecretEMSHostKey:                  fmt.Sprintf("%s", publishURL),
		deployment.PublisherSecretBEBNamespaceKey:  namespace,
	}
}


