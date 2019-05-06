package connectorservice

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509/pkix"
	"fmt"
	"net/http"
	"testing"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates/mocks"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	application   = "test-app"
	tenant        = "test-tenant"
	group         = "test-group"
	eventsURL     = "https://gateway.events/"
	metadataURL   = "https://gateway.metadata/"
	revocationURL = "https://gateway.revocation/"
	renewalURL    = "https://gateway.renewal/"
)

var (
	subject = pkix.Name{
		CommonName: "test-app",
	}

	tlsConfig = &tls.Config{}
)

func TestMutualTLSConnectorClient_GetManagementInfo(t *testing.T) {

	managementInfoEndpoint := "/v1/application/management/info"
	csrProvider := &mocks.CSRProvider{}

	t.Run("should get management info", func(t *testing.T) {
		// given
		server, router := createTestServer()
		defer server.Close()

		router.HandleFunc(managementInfoEndpoint, func(w http.ResponseWriter, r *http.Request) {
			managementInfo := ManagementInfo{
				ClientIdentity: ClientIdentity{
					Application: application,
					Tenant:      tenant,
					Group:       group,
				},
				ManagementURLs: ManagementURLs{
					EventsURL:     eventsURL,
					MetadataURL:   metadataURL,
					RevocationURL: revocationURL,
					RenewalURL:    renewalURL,
				},
			}

			respond(t, w, http.StatusOK, managementInfo)
		})

		mutualTLSClient := NewEstablishedConnectionClient(tlsConfig, csrProvider, subject)

		// when
		managementInfoResponse, err := mutualTLSClient.GetManagementInfo(server.URL + managementInfoEndpoint)

		// then
		require.NoError(t, err)

		assert.Equal(t, application, managementInfoResponse.ClientIdentity.Application)
	})

	t.Run("should return error when request failed", func(t *testing.T) {
		// given
		mutualTLSClient := NewEstablishedConnectionClient(tlsConfig, csrProvider, subject)

		// when
		_, err := mutualTLSClient.GetManagementInfo("https://invalid.url.kyma.cx")

		// then
		require.Error(t, err)
	})

	t.Run("should return error when server responded with error", func(t *testing.T) {
		// given
		server, router := createTestServer()
		defer server.Close()

		router.HandleFunc(managementInfoEndpoint, errorHandler(t))

		mutualTLSClient := NewEstablishedConnectionClient(tlsConfig, csrProvider, subject)

		// when
		_, err := mutualTLSClient.GetManagementInfo(server.URL + managementInfoEndpoint)

		// then
		require.Error(t, err)
	})

}

func TestMutualTLSConnectorClient_RenewCertificate(t *testing.T) {

	encodedCSR := "encodedCSR"
	renewalEndpoint := "/v1/application/certificates/renewals"

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	t.Run("should renew certificate", func(t *testing.T) {
		// given
		csrProvider := &mocks.CSRProvider{}
		csrProvider.On("CreateCSR", subject).Return(encodedCSR, clientKey, nil)

		server, router := createTestServer()
		defer server.Close()

		router.HandleFunc(renewalEndpoint, func(w http.ResponseWriter, r *http.Request) {
			var certRequest CertificateRequest
			err := readResponseBody(r.Body, &certRequest)
			require.NoError(t, err)
			assert.Equal(t, encodedCSR, certRequest.CSR)

			crtResponse := CertificatesResponse{
				CRTChain:  crtChainBase64,
				ClientCRT: clientCRTBase64,
				CaCRT:     caCRTBase64,
			}

			respond(t, w, http.StatusCreated, crtResponse)
		})

		renewalURL := fmt.Sprintf("%s%s", server.URL, renewalEndpoint)

		mutualTLSClient := NewEstablishedConnectionClient(tlsConfig, csrProvider, subject)

		// when
		certificates, err := mutualTLSClient.RenewCertificate(renewalURL)
		require.NoError(t, err)

		// then
		assert.Equal(t, clientCRT, certificates.ClientCRT)
		assert.Equal(t, caCRT, certificates.CaCRT)
		assert.Equal(t, crtChain, certificates.CRTChain)
	})

	t.Run("should return error when failed to create CSR", func(t *testing.T) {
		// given
		csrProvider := &mocks.CSRProvider{}
		csrProvider.On("CreateCSR", subject).Return("", nil, errors.New("error"))

		mutualTLSClient := NewEstablishedConnectionClient(tlsConfig, csrProvider, subject)

		// when
		_, err := mutualTLSClient.RenewCertificate("https://invalid.url.kyma.cx")

		// then
		require.Error(t, err)
	})

	t.Run("should return error when request failed", func(t *testing.T) {
		// given
		csrProvider := &mocks.CSRProvider{}
		csrProvider.On("CreateCSR", subject).Return(encodedCSR, clientKey, nil)

		mutualTLSClient := NewEstablishedConnectionClient(tlsConfig, csrProvider, subject)

		// when
		_, err := mutualTLSClient.RenewCertificate("https://invalid.url.kyma.cx")

		// then
		require.Error(t, err)
	})

	t.Run("should return error when server responded with error", func(t *testing.T) {
		// given
		csrProvider := &mocks.CSRProvider{}
		csrProvider.On("CreateCSR", subject).Return(encodedCSR, clientKey, nil)

		server, router := createTestServer()
		defer server.Close()

		router.HandleFunc(renewalEndpoint, errorHandler(t))

		renewalURL := fmt.Sprintf("%s%s", server.URL, renewalEndpoint)

		mutualTLSClient := NewEstablishedConnectionClient(tlsConfig, csrProvider, subject)

		// when
		_, err := mutualTLSClient.RenewCertificate(renewalURL)

		// then
		require.Error(t, err)
	})
}
