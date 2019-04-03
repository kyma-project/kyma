package connectorservice

import (
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates/mocks"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	csrPath         = "/v1/signingRequests/info"
	certificatePath = "/v1/certificate"

	crtChainBase64  = "Y3J0Q2hhaW4=" // crtChain
	clientCRTBase64 = "Y2xpZW50Q1JU" // clientCRT
	caCRTBase64     = "Y2FDUlQ="     // caCRT

	plainSubject = "OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE,CN=ec-default"

	infoURL = "https://connector-service/v1/runtimes/management/info"
)

var (
	crtChain  = []byte("crtChain")
	clientCRT = []byte("clientCRT")
	caCRT     = []byte("caCRT")
)

func TestConnectorClient_RequestCertificate(t *testing.T) {

	subject := pkix.Name{
		OrganizationalUnit: []string{"OrgUnit"},
		Organization:       []string{"Organization"},
		Locality:           []string{"Waldorf"},
		Province:           []string{"Waldorf"},
		Country:            []string{"DE"},
		CommonName:         "ec-default",
	}

	encodedCSR := "encodedCSR"

	t.Run("should receive certificates", func(t *testing.T) {
		// given
		csrProvider := &mocks.CSRProvider{}
		csrProvider.On("CreateCSR", subject).Return(encodedCSR, nil)

		testServer, router := createTestServer()
		connectorURL := testServer.URL
		defer testServer.Close()

		infoResponse := createInfoResponse(connectorURL)
		router.Handle(csrPath, handler(func(w http.ResponseWriter, r *http.Request) {
			respond(t, w, http.StatusOK, infoResponse)
		}))

		router.Handle(certificatePath, handler(func(w http.ResponseWriter, r *http.Request) {
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
		}))

		csrURL := fmt.Sprintf("%s%s", connectorURL, csrPath)

		connectorClient := NewConnectorClient(csrProvider)

		// when
		connection, err := connectorClient.ConnectToCentralConnector(csrURL)
		require.NoError(t, err)

		// then
		assert.Equal(t, clientCRT, connection.Certificates.ClientCRT)
		assert.Equal(t, caCRT, connection.Certificates.CaCRT)
		assert.Equal(t, crtChain, connection.Certificates.CRTChain)
		assert.Equal(t, infoURL, connection.ManagementInfoURL)
	})

	t.Run("should return error when failed request info response", func(t *testing.T) {
		// given
		testServer, router := createTestServer()
		connectorURL := testServer.URL
		defer testServer.Close()

		router.Handle(csrPath, errorHandler(t))

		csrURL := fmt.Sprintf("%s%s", connectorURL, csrPath)

		connectorClient := NewConnectorClient(nil)

		// when
		_, err := connectorClient.ConnectToCentralConnector(csrURL)
		require.Error(t, err)
	})

	t.Run("should return error when failed to create CSR", func(t *testing.T) {
		// given
		csrProvider := &mocks.CSRProvider{}
		csrProvider.On("CreateCSR", subject).Return("", errors.New("Error"))

		testServer, router := createTestServer()
		connectorURL := testServer.URL
		defer testServer.Close()

		infoResponse := createInfoResponse(connectorURL)
		router.Handle(csrPath, handler(func(w http.ResponseWriter, r *http.Request) {
			respond(t, w, http.StatusOK, infoResponse)
		}))

		csrURL := fmt.Sprintf("%s%s", connectorURL, csrPath)

		connectorClient := NewConnectorClient(csrProvider)

		// when
		_, err := connectorClient.ConnectToCentralConnector(csrURL)
		require.Error(t, err)
	})

	t.Run("should return error when failed get certificate response", func(t *testing.T) {
		// given
		csrProvider := &mocks.CSRProvider{}
		csrProvider.On("CreateCSR", subject).Return(encodedCSR, nil)

		testServer, router := createTestServer()
		connectorURL := testServer.URL
		defer testServer.Close()

		infoResponse := createInfoResponse(connectorURL)
		router.Handle(csrPath, handler(func(w http.ResponseWriter, r *http.Request) {
			respond(t, w, http.StatusOK, infoResponse)
		}))

		router.Handle(certificatePath, errorHandler(t))

		csrURL := fmt.Sprintf("%s%s", connectorURL, csrPath)

		connectorClient := NewConnectorClient(csrProvider)

		// when
		_, err := connectorClient.ConnectToCentralConnector(csrURL)
		require.Error(t, err)
	})

	t.Run("should return error when CSR info url is incorrect", func(t *testing.T) {
		// given
		connectorURL := "https://some-invalid-url.kyma"

		csrURL := fmt.Sprintf("%s%s", connectorURL, csrPath)

		connectorClient := NewConnectorClient(nil)

		// when
		_, err := connectorClient.ConnectToCentralConnector(csrURL)
		require.Error(t, err)
	})

	t.Run("should return error when Certificate URL is incorrect", func(t *testing.T) {
		// given
		csrProvider := &mocks.CSRProvider{}
		csrProvider.On("CreateCSR", subject).Return(encodedCSR, nil)

		testServer, router := createTestServer()
		connectorURL := testServer.URL
		defer testServer.Close()

		infoResponse := createInfoResponse("https://some-invalid-url.kyma")
		router.Handle(csrPath, handler(func(w http.ResponseWriter, r *http.Request) {
			respond(t, w, http.StatusOK, infoResponse)
		}))

		csrURL := fmt.Sprintf("%s%s", connectorURL, csrPath)

		connectorClient := NewConnectorClient(csrProvider)

		// when
		_, err := connectorClient.ConnectToCentralConnector(csrURL)
		require.Error(t, err)
	})

	t.Run("should return error when failed to decode base64", func(t *testing.T) {
		testCases := []CertificatesResponse{
			{CRTChain: "invalid base64"},
			{
				CRTChain:  crtChainBase64,
				ClientCRT: "invalid base64",
			},
			{
				CRTChain:  crtChainBase64,
				ClientCRT: clientCRTBase64,
				CaCRT:     "invalid base64",
			},
		}

		for _, certResponse := range testCases {
			t.Run("should return error when failed to decode certificate", func(t *testing.T) {
				// given
				csrProvider := &mocks.CSRProvider{}
				csrProvider.On("CreateCSR", subject).Return(encodedCSR, nil)

				testServer, router := createTestServer()
				connectorURL := testServer.URL
				defer testServer.Close()

				infoResponse := createInfoResponse(connectorURL)
				router.Handle(csrPath, handler(func(w http.ResponseWriter, r *http.Request) {
					respond(t, w, http.StatusOK, infoResponse)
				}))

				router.Handle(certificatePath, handler(func(w http.ResponseWriter, r *http.Request) {
					var certRequest CertificateRequest
					err := readResponseBody(r.Body, &certRequest)
					require.NoError(t, err)
					assert.Equal(t, encodedCSR, certRequest.CSR)
					respond(t, w, http.StatusCreated, certResponse)
				}))

				csrURL := fmt.Sprintf("%s%s", connectorURL, csrPath)

				connectorClient := NewConnectorClient(csrProvider)

				// when
				_, err := connectorClient.ConnectToCentralConnector(csrURL)
				require.Error(t, err)
			})
		}
	})

}

func createInfoResponse(connectorURL string) InfoResponse {
	return InfoResponse{
		CsrURL: fmt.Sprintf("%s%s", connectorURL, certificatePath),
		CertificateInfo: CertificateInfo{
			Subject: plainSubject,
		},
		Api: APIUrls{
			InfoURL: infoURL,
		},
	}
}

func createTestServer() (*httptest.Server, *mux.Router) {
	router := mux.NewRouter()
	testServer := httptest.NewServer(router)
	return testServer, router
}

func errorHandler(t *testing.T) http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) {
		errorResponse := ErrorResponse{
			Error: "Error",
			Code:  http.StatusInternalServerError,
		}
		respond(t, w, http.StatusInternalServerError, errorResponse)
	})
}

func handler(handlerFunc func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(handlerFunc)
}

func respond(t *testing.T, w http.ResponseWriter, status int, response interface{}) {
	w.WriteHeader(status)

	data, err := json.Marshal(response)
	require.NoError(t, err)

	_, err = w.Write(data)
	require.NoError(t, err)
}
