package connector

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit"

	"github.com/pkg/errors"
)

const (
	rsaKeySize = 4096
)

type Client struct {
	skipVerify bool
	httpClient *http.Client
}

func NewConnectorClient(skipSSLVerify bool) *Client {
	return &Client{
		skipVerify: skipSSLVerify,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLVerify},
			},
		},
	}
}

func (c *Client) EstablishApplicationConnection(infoURL string) (ApplicationConnection, error) {
	csrInfo, err := c.requestSigningRequestInfo(infoURL)
	if err != nil {
		return ApplicationConnection{}, err
	}

	clientKey, certificates, err := c.generateCertificates(csrInfo)
	if err != nil {
		return ApplicationConnection{}, err
	}

	managementInfo, err := c.requestManagementInfo(clientKey, certificates, csrInfo.Api.ManagementInfoURL)
	if err != nil {
		return ApplicationConnection{}, err
	}

	return ApplicationConnection{
		Credentials: ApplicationCredentials{
			ClientKey:    clientKey,
			Certificates: certificates,
		},
		ClientIdentity: managementInfo.ClientIdentity,
		ManagementURLs: managementInfo.URLs,
	}, nil
}

func (c *Client) requestManagementInfo(key *rsa.PrivateKey, certs []*x509.Certificate, managementInfoURL string) (ManagementInfoResponse, error) {
	mtlsClient := testkit.NewMTLSClient(key, certs, c.skipVerify)

	var response *http.Response

	err := testkit.Retry(testkit.DefaultRetryConfig, func() (bool, error) {
		req, err := http.NewRequest(http.MethodGet, managementInfoURL, nil)
		if err != nil {
			return true, nil
		}

		response, err = mtlsClient.Do(req)
		if err != nil {
			return true, nil
		}

		return false, nil
	})
	defer closeResponseBody(response)

	if response.StatusCode != http.StatusOK {
		responseError := errors.New(fmt.Sprintf("Failed to fetch Management Info. Received status: %s", response.Status))
		return ManagementInfoResponse{}, dumpErrorResponse(responseError, response)
	}

	var managementInfo ManagementInfoResponse
	err = json.NewDecoder(response.Body).Decode(&managementInfo)
	if err != nil {
		return ManagementInfoResponse{}, errors.Wrap(err, "Failed to decode Management Info Response")
	}

	return managementInfo, nil
}

func (c *Client) generateCertificates(csrInfo CSRInfoResponse) (*rsa.PrivateKey, []*x509.Certificate, error) {
	clientKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to generate key")
	}

	csr, err := createCSR(csrInfo.Certificate.Subject, clientKey)
	if err != nil {
		return nil, nil, err
	}

	certificateResponse, err := c.requestCertificate(csrInfo.CertUrl, csr)
	if err != nil {
		return nil, nil, err
	}

	certificates, err := decodeCertificateResponse(certificateResponse)
	if err != nil {
		return nil, nil, err
	}

	return clientKey, certificates, nil
}

func (c *Client) requestSigningRequestInfo(signingRequestInfoURL string) (CSRInfoResponse, error) {
	var response *http.Response

	err := testkit.Retry(testkit.DefaultRetryConfig, func() (bool, error) {
		req, err := http.NewRequest(http.MethodGet, signingRequestInfoURL, nil)
		if err != nil {
			return true, nil
		}

		response, err = c.httpClient.Do(req)
		if err != nil {
			return true, nil
		}

		return false, nil
	})
	defer closeResponseBody(response)

	if response.StatusCode != http.StatusOK {
		responseError := errors.New(fmt.Sprintf("Failed to fetch fetch CSR Info Response. Received status: %s", response.Status))
		return CSRInfoResponse{}, dumpErrorResponse(responseError, response)
	}

	var csrInfoResponse CSRInfoResponse
	err = json.NewDecoder(response.Body).Decode(&csrInfoResponse)
	if err != nil {
		return CSRInfoResponse{}, errors.Wrap(err, "Failed to decode CSR Info Response")
	}

	return csrInfoResponse, nil
}

func (c *Client) requestCertificate(certificatesURL string, csr []byte) (CrtResponse, error) {
	certRequest := CsrRequest{Csr: base64.StdEncoding.EncodeToString(csr)}
	payload, err := json.Marshal(certRequest)
	if err != nil {
		return CrtResponse{}, err
	}

	var response *http.Response

	err = testkit.Retry(testkit.DefaultRetryConfig, func() (bool, error) {
		req, err := http.NewRequest(http.MethodPost, certificatesURL, bytes.NewBuffer(payload))
		if err != nil {
			return true, nil
		}

		response, err = c.httpClient.Do(req)
		if err != nil {
			return true, nil
		}

		return false, nil
	})

	defer closeResponseBody(response)

	if response.StatusCode != http.StatusCreated {
		responseError := errors.New(fmt.Sprintf("Failed to fetch Certificates. Received status: %s", response.Status))
		return CrtResponse{}, dumpErrorResponse(responseError, response)
	}

	var certificateResponse CrtResponse
	err = json.NewDecoder(response.Body).Decode(&certificateResponse)

	return certificateResponse, nil
}

func decodeCertificateResponse(certResponse CrtResponse) ([]*x509.Certificate, error) {
	rawPem, err := base64.StdEncoding.DecodeString(certResponse.CRTChain)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode certificate chain")
	}

	var certificates []*x509.Certificate

	for block, rest := pem.Decode(rawPem); block != nil; {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse certificate")
		}

		certificates = append(certificates, cert)

		block, rest = pem.Decode(rest)
	}

	if len(certificates) < 2 {
		return nil, errors.New(fmt.Sprintf("Certificate response does not contain enough certificates, found: %d", len(certificates)))
	}

	return certificates, nil
}

func createCSR(strSubject string, keys *rsa.PrivateKey) ([]byte, error) {
	subject := ParseSubject(strSubject)

	var csrTemplate = x509.CertificateRequest{
		Subject: subject,
	}

	certificateRequest, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create CSR")
	}

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: certificateRequest,
	})

	return csr, nil
}

func closeResponseBody(response *http.Response) {
	if response != nil {
		response.Body.Close()
	}
}

func ParseSubject(subject string) pkix.Name {
	subjectInfo := extractSubject(subject)

	return pkix.Name{
		CommonName:         subjectInfo["CN"],
		Country:            []string{subjectInfo["C"]},
		Organization:       []string{subjectInfo["O"]},
		OrganizationalUnit: []string{subjectInfo["OU"]},
		Locality:           []string{subjectInfo["L"]},
		Province:           []string{subjectInfo["ST"]},
	}
}

func extractSubject(subject string) map[string]string {
	result := map[string]string{}

	segments := strings.Split(subject, ",")

	for _, segment := range segments {
		parts := strings.Split(segment, "=")
		result[parts[0]] = parts[1]
	}

	return result
}

func dumpErrorResponse(responseError error, response *http.Response) error {
	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		return errors.WithMessagef(responseError, "Failed to dump response: %s", err.Error())
	}

	return errors.WithMessagef(responseError, "Response: %s", string(dump))
}
