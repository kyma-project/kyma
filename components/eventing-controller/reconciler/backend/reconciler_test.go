package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
)

func TestComputeSecretForPublisher(t *testing.T) {
	testCases := []struct {
		name           string
		messagingData  []byte
		namespaceData  []byte
		expectedSecret v1.Secret
		expectedError  error
	}{
		{
			messagingData: []byte("W3sJCSJicm9rZXIiOiB7CQkJInR5cGUiOiAic2FwbWd3IgkJfSwJCSJvYTIiOiB7CQkJImNsaWVudGlkIjogImNsaWVudGlkIiwJCQkiY2xpZW50c2VjcmV0IjogImNsaWVudHNlY3JldCIsCQkJImdyYW50dHlwZSI6ICJjbGllbnRfY3JlZGVudGlhbHMiLAkJCSJ0b2tlbmVuZHBvaW50IjogImh0dHBzOi8vdG9rZW4iCQl9LAkJInByb3RvY29sIjogWyJhbXFwMTB3cyJdLAkJInVyaSI6ICJ3c3M6Ly9hbXFwIgl9LCB7CQkiYnJva2VyIjogewkJCSJ0eXBlIjogInNhcG1ndyIJCX0sCQkib2EyIjogewkJCSJjbGllbnRpZCI6ICJjbGllbnRpZCIsCQkJImNsaWVudHNlY3JldCI6ICJjbGllbnRzZWNyZXQiLAkJCSJncmFudHR5cGUiOiAiY2xpZW50X2NyZWRlbnRpYWxzIiwJCQkidG9rZW5lbmRwb2ludCI6ICJodHRwczovL3Rva2VuIgkJfSwJCSJwcm90b2NvbCI6IFsiYW1xcDEwd3MiXSwJCSJ1cmkiOiAid3NzOi8vYW1xcCIJfSwJewkJImJyb2tlciI6IHsJCQkidHlwZSI6ICJzYXByZXN0bWd3IgkJfSwJCSJvYTIiOiB7CQkJImNsaWVudGlkIjogInJlc3QtY2xpZW50aWQiLAkJCSJjbGllbnRzZWNyZXQiOiAicmVzdC1jbGllbnQtc2VjcmV0IiwJCQkiZ3JhbnR0eXBlIjogImNsaWVudF9jcmVkZW50aWFscyIsCQkJInRva2VuZW5kcG9pbnQiOiAiaHR0cHM6Ly9yZXN0LXRva2VuIgkJfSwJCSJwcm90b2NvbCI6IFsiaHR0cHJlc3QiXSwJCSJ1cmkiOiAiaHR0cHM6Ly9yZXN0LW1lc3NhZ2luZyIJfV0="),
			namespaceData: []byte("dmFsaWQvbmFtZXNwYWNl"),
			expectedSecret: v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      PublisherName,
					Namespace: PublisherNamespace,
					Labels: map[string]string{
						AppLabelKey: PublisherName,
					},
				},
				StringData: map[string]string{
					"client-id":       "rest-clientid",
					"client-secret":   "rest-client-secret",
					"token-endpoint":  "https://rest-token?grant_type=client_credentials&response_type=token",
					"ems-publish-url": "https://rest-messaging",
					"beb-namespace":   "valid/namespace",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			publisherSecret := v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: tc.name,
				},
				Data: map[string][]byte{
					"messaging": tc.messagingData,
					"namespace": tc.namespaceData,
				},
			}

			gotPublisherSecret, err := computeSecretForPublisher(&publisherSecret)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedSecret, *gotPublisherSecret, "invalid publisher secret")
		})
	}
}
