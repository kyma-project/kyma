package connector

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
)

const (
	tokenEndpoint = "/v1/applications/tokens"

	ApplicationHeader = "Application"
	GroupHeader       = "Group"
	TenantHeader      = "Tenant"

	rsaKeySize = 4096
)

type Client struct {
	httpClient                  *http.Client
	connectorServiceInternalURL string
}

// TODO - possibly do it with Connection Token Handler
func NewConnectorClient(connectorServiceInternalURL string) *Client {
	return &Client{
		httpClient:                  &http.Client{},
		connectorServiceInternalURL: connectorServiceInternalURL,
	}
}

func (c *Client) EstablishApplicationConnection(application *v1alpha1.Application) (ApplicationConnection, error) {
	tokenResponse, err := c.requestToken(application)
	if err != nil {
		return ApplicationConnection{}, err
	}

	csrInfo, err := c.requestSigningRequestInfo(tokenResponse.URL)
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
		ClientKey:      clientKey,
		Certificates:   certificates,
		ClientIdentity: managementInfo.ClientIdentity,
		ManagementURLs: managementInfo.URLs,
	}, nil
}

func (c *Client) requestManagementInfo(key *rsa.PrivateKey, certs []*x509.Certificate, managementInfoURL string) (ManagementInfoResponse, error) {
	mtlsClient := testkit.NewMTLSClient(key, certs)

	req, err := http.NewRequest(http.MethodGet, managementInfoURL, nil)
	if err != nil {
		return ManagementInfoResponse{}, err
	}

	response, err := mtlsClient.Do(req)
	if err != nil {
		return ManagementInfoResponse{}, errors.Wrap(err, "Failed to fetch Management Info")
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

func (c *Client) requestToken(application *v1alpha1.Application) (TokenResponse, error) {
	req, err := http.NewRequest(http.MethodPost, c.connectorServiceInternalURL+tokenEndpoint, nil)
	if err != nil {
		return TokenResponse{}, err
	}
	req.Header.Set(ApplicationHeader, application.Name)

	if application.Spec.HasGroup() && application.Spec.HasTenant() {
		req.Header.Set(TenantHeader, application.Spec.Tenant)
		req.Header.Set(GroupHeader, application.Spec.Group)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return TokenResponse{}, errors.Wrap(err, "Failed to fetch token")
	}
	defer response.Body.Close()

	var tokenResponse TokenResponse
	err = json.NewDecoder(response.Body).Decode(&tokenResponse)
	if err != nil {
		return TokenResponse{}, errors.Wrap(err, "Failed to decode token response")
	}

	return tokenResponse, nil
}

func (c *Client) requestSigningRequestInfo(signingRequestInfoURL string) (CSRInfoResponse, error) {
	req, err := http.NewRequest(http.MethodGet, signingRequestInfoURL, nil)
	if err != nil {
		return CSRInfoResponse{}, err
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return CSRInfoResponse{}, errors.Wrap(err, "Failed to fetch CSR Info Response")
	}
	defer response.Body.Close()

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

	req, err := http.NewRequest(http.MethodPost, certificatesURL, bytes.NewBuffer(payload))
	if err != nil {
		return CrtResponse{}, err
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return CrtResponse{}, err
	}
	defer response.Body.Close()

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
		rawPem = rest
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
