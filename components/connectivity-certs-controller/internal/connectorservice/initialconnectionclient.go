package connectorservice

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"

	"github.com/pkg/errors"
)

type InitialConnectionClient interface {
	Establish(csrInfoURL string) (EstablishedConnection, error)
}

type initialConnectionClient struct {
	csrProvider certificates.CSRProvider
	httpClient  *http.Client
}

func NewInitialConnectionClient(csrProvider certificates.CSRProvider) InitialConnectionClient {
	return &initialConnectionClient{
		httpClient:  &http.Client{},
		csrProvider: csrProvider,
	}
}

func (cc *initialConnectionClient) Establish(csrInfoURL string) (EstablishedConnection, error) {
	infoResponse, err := cc.requestCSRInfo(csrInfoURL)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed while requesting CSR info")
	}

	subject := parseSubject(infoResponse.CertificateInfo.Subject)

	encodedCSR, clientKey, err := cc.csrProvider.CreateCSR(subject)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed while creating CSR")
	}

	certificateResponse, err := cc.requestCertificates(infoResponse.CsrURL, encodedCSR)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed while requesting certificate")
	}

	return composeConnectionData(clientKey, certificateResponse, infoResponse.Api.InfoURL)
}

func (cc *initialConnectionClient) requestCSRInfo(csrInfoURL string) (InfoResponse, error) {
	csrInfoReq, err := http.NewRequest(http.MethodGet, csrInfoURL, nil)
	if err != nil {
		return InfoResponse{}, err
	}

	response, err := cc.httpClient.Do(csrInfoReq)
	if err != nil {
		return InfoResponse{}, err
	}

	if response.StatusCode != http.StatusOK {
		message := fmt.Sprintf("Connector Service CSR info responded with unexpected status %d", response.StatusCode)
		return InfoResponse{}, errors.Wrap(extractErrorResponse(response), message)
	}

	var infoResponse InfoResponse
	err = readResponseBody(response.Body, &infoResponse)
	if err != nil {
		return InfoResponse{}, err
	}

	return infoResponse, nil
}

func (cc *initialConnectionClient) requestCertificates(csrURL string, encodedCSR string) (CertificatesResponse, error) {
	requestData, err := json.Marshal(CertificateRequest{CSR: encodedCSR})
	if err != nil {
		return CertificatesResponse{}, err
	}

	certificateReq, err := http.NewRequest(http.MethodPost, csrURL, bytes.NewBuffer(requestData))
	if err != nil {
		return CertificatesResponse{}, err
	}

	response, err := cc.httpClient.Do(certificateReq)
	if err != nil {
		return CertificatesResponse{}, err
	}

	if response.StatusCode != http.StatusCreated {
		message := fmt.Sprintf("Connector Service Certificates responded with unexpected status %d", response.StatusCode)
		return CertificatesResponse{}, errors.Wrap(extractErrorResponse(response), message)
	}

	var certificateResponse CertificatesResponse
	err = readResponseBody(response.Body, &certificateResponse)
	if err != nil {
		return CertificatesResponse{}, err
	}

	return certificateResponse, nil
}

func composeConnectionData(clientKey *rsa.PrivateKey, certificateResponse CertificatesResponse, managementInfoURL string) (EstablishedConnection, error) {
	certs, err := decodeCertificateResponse(clientKey, certificateResponse)
	if err != nil {
		return EstablishedConnection{}, err
	}

	return EstablishedConnection{
		Certificates:      certs,
		ManagementInfoURL: managementInfoURL,
	}, nil
}

func decodeCertificateResponse(clientKey *rsa.PrivateKey, certificateResponse CertificatesResponse) (certificates.Certificates, error) {
	crtChain, err := base64.StdEncoding.DecodeString(certificateResponse.CRTChain)
	if err != nil {
		return certificates.Certificates{}, errors.Wrap(err, "Failed to decode certificate chain")
	}

	clientCRT, err := base64.StdEncoding.DecodeString(certificateResponse.ClientCRT)
	if err != nil {
		return certificates.Certificates{}, errors.Wrap(err, "Failed to decode client certificate")
	}

	caCRT, err := base64.StdEncoding.DecodeString(certificateResponse.CaCRT)
	if err != nil {
		return certificates.Certificates{}, errors.Wrap(err, "Failed to decode CA certificate")
	}

	return certificates.Certificates{
		ClientKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)}),
		CRTChain:  crtChain,
		ClientCRT: clientCRT,
		CaCRT:     caCRT,
	}, nil
}

func extractErrorResponse(response *http.Response) error {
	bodyReader1, bodyReader2, err := drainBody(response.Body)
	if err != nil {
		return errors.New("Failed to read error response")
	}

	var errorResponse ErrorResponse
	err = unmarshalFromReader(bodyReader1, &errorResponse)
	if err != nil {
		response.Body = bodyReader2
		dump, err := httputil.DumpResponse(response, true)
		if err != nil {
			return errors.Wrap(err, "Failed to dump error response")
		}

		errorDump := fmt.Sprintf("Error response: %s", string(dump))
		return errors.New(errorDump)
	}

	return errors.New(errorResponse.Error)
}

func readResponseBody(body io.ReadCloser, model interface{}) error {
	defer body.Close()
	return unmarshalFromReader(body, model)
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func unmarshalFromReader(body io.Reader, model interface{}) error {
	rawData, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	return json.Unmarshal(rawData, &model)
}

func parseSubject(plainSubject string) pkix.Name {
	subjectInfo := extractSubject(plainSubject)

	return pkix.Name{
		CommonName:         subjectInfo["CN"],
		Country:            []string{subjectInfo["C"]},
		Organization:       []string{subjectInfo["O"]},
		OrganizationalUnit: []string{subjectInfo["OU"]},
		Locality:           []string{subjectInfo["L"]},
		Province:           []string{subjectInfo["ST"]},
	}
}

func extractSubject(plainSubject string) map[string]string {
	result := map[string]string{}

	segments := strings.Split(plainSubject, ",")

	for _, segment := range segments {
		parts := strings.Split(segment, "=")
		result[parts[0]] = parts[1]
	}

	return result
}
