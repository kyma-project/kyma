package backend

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"

	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
)

// TestGetSecretForPublisher verifies the successful and failing retrieval
// of secrets.
func TestGetSecretForPublisher(t *testing.T) {
	secretFor := func(message, namespace []byte) *corev1.Secret {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: deployment.PublisherName,
			},
		}

		secret.Data = make(map[string][]byte)

		if len(message) > 0 {
			secret.Data["messaging"] = message
		}
		if len(namespace) > 0 {
			secret.Data["namespace"] = namespace
		}

		return secret
	}

	testCases := []struct {
		name           string
		messagingData  []byte
		namespaceData  []byte
		expectedSecret corev1.Secret
		expectedError  error
	}{
		{
			name:          "with valid message and namespace data",
			messagingData: []byte("[{		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	}, {		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	},	{		\"broker\": {			\"type\": \"saprestmgw\"		},		\"oa2\": {			\"clientid\": \"rest-clientid\",			\"clientsecret\": \"rest-client-secret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://rest-token\"		},		\"protocol\": [\"httprest\"],		\"uri\": \"https://rest-messaging\"	}]"),
			namespaceData: []byte("valid/namespace"),
			expectedSecret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deployment.PublisherName,
					Namespace: deployment.PublisherNamespace,
					Labels: map[string]string{
						deployment.AppLabelKey: deployment.PublisherName,
					},
				},
				StringData: map[string]string{
					"client-id":        "rest-clientid",
					"client-secret":    "rest-client-secret",
					"token-endpoint":   "https://rest-token?grant_type=client_credentials&response_type=token",
					"ems-publish-host": "https://rest-messaging",
					"ems-publish-url":  "https://rest-messaging/sap/ems/v1/events",
					"beb-namespace":    "valid/namespace",
				},
			},
		},
		{
			name:          "with empty message data",
			namespaceData: []byte("valid/namespace"),
			expectedError: errors.New("message is missing from BEB secret"),
		},
		{
			name:          "with empty namespace data", // nolint:gofmt
			messagingData: []byte("[{		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	}, {		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	},	{		\"broker\": {			\"type\": \"saprestmgw\"		},		\"oa2\": {			\"clientid\": \"rest-clientid\",			\"clientsecret\": \"rest-client-secret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://rest-token\"		},		\"protocol\": [\"httprest\"],		\"uri\": \"https://rest-messaging\"	}]"),
			expectedError: errors.New("namespace is missing from BEB secret"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			publisherSecret := secretFor(tc.messagingData, tc.namespaceData)

			gotPublisherSecret, err := getSecretForPublisher(publisherSecret)
			if tc.expectedError != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error(), "invalid error")
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedSecret, *gotPublisherSecret, "invalid publisher secret")
		})
	}
}

func Test_updateMutatingValidatingWebhookWithCABundle(t *testing.T) {
	// given
	ctx := context.Background()
	dummyCABundle := make([]byte, 20)
	rand.Read(dummyCABundle)
	newCABundle := make([]byte, 20)
	rand.Read(newCABundle)

	testCases := []struct {
		name             string
		givenObjects     []client.Object
		wantMutatingWH   *admissionv1.MutatingWebhookConfiguration
		wantValidatingWH *admissionv1.ValidatingWebhookConfiguration
		wantError        error
	}{
		{
			name:      "secret does not exist",
			wantError: errObjectNotFound,
		},
		{
			name: "secret exits but mutatingWH does not exist",
			givenObjects: []client.Object{
				getSecretWithTLSSecret(nil),
			},
			wantError: errObjectNotFound,
		},
		{
			name: "mutatingWH exists, validatingWH does not exist",
			givenObjects: []client.Object{
				getSecretWithTLSSecret(nil),
				getMutatingWebhookConfig([]admissionv1.MutatingWebhook{
					{
						ClientConfig: admissionv1.WebhookClientConfig{},
					},
				}),
			},
			wantError: errObjectNotFound,
		},
		{
			name: "mutatingWH, validatingWH exists but does not contain webhooks",
			givenObjects: []client.Object{
				getSecretWithTLSSecret(nil),
				getMutatingWebhookConfig(nil),
				getValidatingWebhookConfig(nil),
			},
			wantError: errInvalidObject,
		},
		{
			name: "validatingWH does not contain webhooks",
			givenObjects: []client.Object{
				getSecretWithTLSSecret(nil),
				getMutatingWebhookConfig([]admissionv1.MutatingWebhook{
					{
						ClientConfig: admissionv1.WebhookClientConfig{},
					},
				}),
				getValidatingWebhookConfig(nil),
			},
			wantError: errInvalidObject,
		},
		{
			name: "WH does not contain valid CABundle",
			givenObjects: []client.Object{
				getSecretWithTLSSecret(dummyCABundle),
				getMutatingWebhookConfig([]admissionv1.MutatingWebhook{
					{
						ClientConfig: admissionv1.WebhookClientConfig{},
					},
				}),
				getValidatingWebhookConfig([]admissionv1.ValidatingWebhook{
					{
						ClientConfig: admissionv1.WebhookClientConfig{},
					},
				}),
			},
			wantMutatingWH: getMutatingWebhookConfig([]admissionv1.MutatingWebhook{
				{
					ClientConfig: admissionv1.WebhookClientConfig{
						CABundle: dummyCABundle,
					},
				},
			}),
			wantValidatingWH: getValidatingWebhookConfig([]admissionv1.ValidatingWebhook{
				{
					ClientConfig: admissionv1.WebhookClientConfig{
						CABundle: dummyCABundle,
					},
				},
			}),
			wantError: nil,
		},
		{
			name: "WH contains valid CABundle",
			givenObjects: []client.Object{
				getSecretWithTLSSecret(dummyCABundle),
				getMutatingWebhookConfig([]admissionv1.MutatingWebhook{
					{
						ClientConfig: admissionv1.WebhookClientConfig{
							CABundle: dummyCABundle,
						},
					},
				}),
				getValidatingWebhookConfig([]admissionv1.ValidatingWebhook{
					{
						ClientConfig: admissionv1.WebhookClientConfig{
							CABundle: dummyCABundle,
						},
					},
				}),
			},
			wantMutatingWH: getMutatingWebhookConfig([]admissionv1.MutatingWebhook{
				{
					ClientConfig: admissionv1.WebhookClientConfig{
						CABundle: dummyCABundle,
					},
				},
			}),
			wantValidatingWH: getValidatingWebhookConfig([]admissionv1.ValidatingWebhook{
				{
					ClientConfig: admissionv1.WebhookClientConfig{
						CABundle: dummyCABundle,
					},
				},
			}),
			wantError: nil,
		},
		{
			name: "WH contains outdated valid CABundle",
			givenObjects: []client.Object{
				getSecretWithTLSSecret(newCABundle),
				getMutatingWebhookConfig([]admissionv1.MutatingWebhook{
					{
						ClientConfig: admissionv1.WebhookClientConfig{
							CABundle: dummyCABundle,
						},
					},
				}),
				getValidatingWebhookConfig([]admissionv1.ValidatingWebhook{
					{
						ClientConfig: admissionv1.WebhookClientConfig{
							CABundle: dummyCABundle,
						},
					},
				}),
			},
			wantMutatingWH: getMutatingWebhookConfig([]admissionv1.MutatingWebhook{
				{
					ClientConfig: admissionv1.WebhookClientConfig{
						CABundle: newCABundle,
					},
				},
			}),
			wantValidatingWH: getValidatingWebhookConfig([]admissionv1.ValidatingWebhook{
				{
					ClientConfig: admissionv1.WebhookClientConfig{
						CABundle: newCABundle,
					},
				},
			}),
			wantError: nil,
		},
	}

	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			// given
			r := setup(tC.givenObjects...)

			// when
			err := r.updateMutatingValidatingWebhookWithCABundle(ctx)

			// then
			require.ErrorIs(t, err, tC.wantError)
			if tC.wantError == nil {
				mutatingWH, validatingWH, newErr := r.getMutatingAndValidatingWebHookConfig(ctx)
				require.NoError(t, newErr)
				require.Equal(t, mutatingWH.Webhooks[0], tC.wantMutatingWH.Webhooks[0])
				require.Equal(t, validatingWH.Webhooks[0], tC.wantValidatingWH.Webhooks[0])
			}
		})
	}
}

func setup(objs ...client.Object) Reconciler {
	fakeClient := fake.NewClientBuilder().WithObjects(objs...).Build()
	return Reconciler{
		Client: fakeClient,
		cfg:    getTestBackendConfig(),
	}
}

func getSecretWithTLSSecret(dummyCABundle []byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getTestBackendConfig().WebhookSecretName,
			Namespace: certificateSecretNamespace,
		},
		Data: map[string][]byte{
			tlsCertField: dummyCABundle,
		},
	}
}

func getMutatingWebhookConfig(webhook []admissionv1.MutatingWebhook) *admissionv1.MutatingWebhookConfiguration {
	return &admissionv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: getTestBackendConfig().MutatingWebhookName,
		},
		Webhooks: webhook,
	}
}

func getValidatingWebhookConfig(webhook []admissionv1.ValidatingWebhook) *admissionv1.ValidatingWebhookConfiguration {
	return &admissionv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: getTestBackendConfig().ValidatingWebhookName,
		},
		Webhooks: webhook,
	}
}

func getTestBackendConfig() env.BackendConfig {
	return env.BackendConfig{
		WebhookSecretName:     "webhookSecret",
		MutatingWebhookName:   "mutatingWH",
		ValidatingWebhookName: "validatingWH",
	}
}
