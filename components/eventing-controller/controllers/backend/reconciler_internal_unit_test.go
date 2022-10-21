package backend

import (
	"errors"
	"testing"

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
