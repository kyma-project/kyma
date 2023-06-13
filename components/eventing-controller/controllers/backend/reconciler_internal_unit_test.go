package backend

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/internal/featureflags"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
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
			name:          "with empty namespace data",
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

// TestGetSecretForPublisher verifies the successful and failing retrieval
// of secrets.
func TestCreateDeleteNATSSecret(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		givenSecret *corev1.Secret
	}{
		{
			name:        "create secret when secret does not exist",
			givenSecret: &corev1.Secret{},
		},
		{
			name:        "do not recreate secret when secret exists",
			givenSecret: constructNATSSecret(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			r := setup(tc.givenSecret)

			// when
			err := r.createNATSSecret(ctx)

			// then
			assert.NoError(t, err)
			gotSecret := corev1.Secret{}
			err = r.Client.Get(ctx, client.ObjectKey{Name: natsSecretName, Namespace: kymaSystemNamespace}, &gotSecret)
			assert.NoError(t, err)

			if tc.givenSecret.ObjectMeta.Name == natsSecretName {
				assert.Equal(t, gotSecret.Data, tc.givenSecret.Data)
			} else {
				assert.NotNil(t, gotSecret.Data)
			}

			// when
			err = r.deleteNATSSecret(ctx)

			// then
			assert.NoError(t, err)
			gotSecret = corev1.Secret{}
			err = r.Client.Get(ctx, client.ObjectKey{Name: natsSecretName, Namespace: kymaSystemNamespace}, &gotSecret)
			assert.True(t, k8serrors.IsNotFound(err))
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

func Test_getOAuth2ClientCredentials(t *testing.T) {
	const (
		defaultWebhookTokenEndpoint               = "http://domain.com/token"
		defaultEventingWebhookAuthSecretName      = "eventing-webhook-auth"
		defaultEventingWebhookAuthSecretNamespace = "kyma-system"
	)

	testCases := []struct {
		name             string
		givenFlagEnabled bool
		givenSecrets     []*corev1.Secret
		wantError        bool
		wantClientID     []byte
		wantClientSecret []byte
		wantTokenURL     []byte
		wantCertsURL     []byte
	}{
		// secret is not found
		{
			name:             "eventing auth manager enabled, but secret does not exist",
			givenFlagEnabled: true,
			givenSecrets:     nil,
			wantError:        true,
		},
		{
			name:             "eventing auth manager disabled, but secret does not exist",
			givenFlagEnabled: false,
			givenSecrets:     nil,
			wantError:        true,
		},
		// secret is found but some of the required data is missing
		{
			name:             "eventing auth manager enabled, and secret exists with missing data",
			givenFlagEnabled: true,
			givenSecrets: []*corev1.Secret{
				// required secret
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaultEventingWebhookAuthSecretName,
						Namespace: defaultEventingWebhookAuthSecretNamespace,
					},
					Data: map[string][]byte{
						secretKeyClientID: []byte("test-client-id-0"),
						// missing data
					},
				},
				// not required secret
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      getOAuth2ClientSecretName(),
						Namespace: deployment.ControllerNamespace,
					},
					Data: map[string][]byte{
						secretKeyClientID:     []byte("test-client-id-1"),
						secretKeyClientSecret: []byte("test-client-secret-1"),
						secretKeyTokenURL:     []byte("test-token-url-1"),
						secretKeyCertsURL:     []byte("test-certs-url-1"),
					},
				},
			},
			wantError: true,
		},
		{
			name:             "eventing auth manager disabled, and secret exists with missing data",
			givenFlagEnabled: false,
			givenSecrets: []*corev1.Secret{
				// not required secret
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaultEventingWebhookAuthSecretName,
						Namespace: defaultEventingWebhookAuthSecretNamespace,
					},
					Data: map[string][]byte{
						secretKeyClientID:     []byte("test-client-id-0"),
						secretKeyClientSecret: []byte("test-client-secret-0"),
						secretKeyTokenURL:     []byte("test-token-url-0"),
						secretKeyCertsURL:     []byte("test-certs-url-0"),
					},
				},
				// required secret
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      getOAuth2ClientSecretName(),
						Namespace: deployment.ControllerNamespace,
					},
					Data: map[string][]byte{
						secretKeyClientID: []byte("test-client-id-1"),
						// missing data
					},
				},
			},
			wantError: true,
		},
		// secret is found with all the required data
		{
			name:             "eventing auth manager enabled, and secret exists with all data",
			givenFlagEnabled: true,
			givenSecrets: []*corev1.Secret{
				// required secret
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaultEventingWebhookAuthSecretName,
						Namespace: defaultEventingWebhookAuthSecretNamespace,
					},
					Data: map[string][]byte{
						secretKeyClientID:     []byte("test-client-id-0"),
						secretKeyClientSecret: []byte("test-client-secret-0"),
						secretKeyTokenURL:     []byte("test-token-url-0"),
						secretKeyCertsURL:     []byte("test-certs-url-0"),
					},
				},
				// not required secret
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      getOAuth2ClientSecretName(),
						Namespace: deployment.ControllerNamespace,
					},
					Data: map[string][]byte{
						secretKeyClientID:     []byte("test-client-id-1"),
						secretKeyClientSecret: []byte("test-client-secret-1"),
						secretKeyTokenURL:     []byte("test-token-url-1"),
						secretKeyCertsURL:     []byte("test-certs-url-1"),
					},
				},
			},
			wantError:        false,
			wantClientID:     []byte("test-client-id-0"),
			wantClientSecret: []byte("test-client-secret-0"),
			wantTokenURL:     []byte("test-token-url-0"),
			wantCertsURL:     []byte("test-certs-url-0"),
		},
		{
			name:             "eventing auth manager disabled, and secret exists with all data",
			givenFlagEnabled: false,
			givenSecrets: []*corev1.Secret{
				// not required secret
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaultEventingWebhookAuthSecretName,
						Namespace: defaultEventingWebhookAuthSecretNamespace,
					},
					Data: map[string][]byte{
						secretKeyClientID:     []byte("test-client-id-0"),
						secretKeyClientSecret: []byte("test-client-secret-0"),
						secretKeyTokenURL:     []byte("test-token-url-0"),
						secretKeyCertsURL:     []byte("test-certs-url-0"),
					},
				},
				// required secret
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      getOAuth2ClientSecretName(),
						Namespace: deployment.ControllerNamespace,
					},
					Data: map[string][]byte{
						secretKeyClientID:     []byte("test-client-id-1"),
						secretKeyClientSecret: []byte("test-client-secret-1"),
						secretKeyTokenURL:     []byte("test-token-url-1"),
						secretKeyCertsURL:     []byte("test-certs-url-1"),
					},
				},
			},
			wantError:        false,
			wantClientID:     []byte("test-client-id-1"),
			wantClientSecret: []byte("test-client-secret-1"),
			wantTokenURL:     []byte(defaultWebhookTokenEndpoint),
			wantCertsURL:     []byte(""),
		},
	}

	l, e := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, e)

	for _, testcase := range testCases {
		tc := testcase
		t.Run(tc.name, func(t *testing.T) {
			// given
			ctx := context.Background()
			featureflags.SetEventingWebhookAuthEnabled(tc.givenFlagEnabled)
			r := Reconciler{
				Client: fake.NewClientBuilder().WithObjects().Build(),
				logger: l,
				envCfg: env.Config{
					WebhookTokenEndpoint: defaultWebhookTokenEndpoint,
				},
				cfg: env.BackendConfig{
					EventingWebhookAuthSecretName:      defaultEventingWebhookAuthSecretName,
					EventingWebhookAuthSecretNamespace: defaultEventingWebhookAuthSecretNamespace,
				},
			}
			if len(tc.givenSecrets) > 0 {
				for _, secret := range tc.givenSecrets {
					err := r.Client.Create(ctx, secret)
					require.NoError(t, err)
				}
			}

			// when
			credentials, err := r.getOAuth2ClientCredentials(ctx)

			// then
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.Equal(t, tc.wantClientID, credentials.clientID)
			require.Equal(t, tc.wantClientSecret, credentials.clientSecret)
			require.Equal(t, tc.wantTokenURL, credentials.tokenURL)
			require.Equal(t, tc.wantCertsURL, credentials.certsURL)
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
			Namespace: kymaSystemNamespace,
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
