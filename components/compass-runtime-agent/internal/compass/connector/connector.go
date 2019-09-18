package connector

import (
	"crypto/x509/pkix"
	"strings"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
)

type EstablishedConnection struct {
	Credentials certificates.Credentials
	DirectorURL string
}

//go:generate mockery -name=Connector
type Connector interface {
	EstablishConnection(token string) (EstablishedConnection, error)
}

func NewCompassConnector(connectorURL string, csrProvider certificates.CSRProvider, tokenSecuredConnectorClient TokenSecuredClient) Connector {
	return &compassConnector{
		connectorURL:                connectorURL,
		csrProvider:                 csrProvider,
		tokenSecuredConnectorClient: tokenSecuredConnectorClient,
		//certSecuredConnectorClient:  certSecuredConnectorClient, // TODO - cert secured client constructor?
	}
}

type compassConnector struct {
	connectorURL                string
	csrProvider                 certificates.CSRProvider
	tokenSecuredConnectorClient TokenSecuredClient
	certSecuredConnectorClient  CertificateSecuredClient
}

func (cc *compassConnector) EstablishConnection(token string) (EstablishedConnection, error) {
	if cc.connectorURL == "" {
		return EstablishedConnection{}, errors.New("Failed to establish connection. Connector URL is empty")
	}

	configuration, err := cc.tokenSecuredConnectorClient.Configuration(token)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to fetch configuration")
	}

	subject := parseSubject(configuration.CertificateSigningRequestInfo.Subject)
	csr, key, err := cc.csrProvider.CreateCSR(subject)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to generate CSR")
	}

	certResponse, err := cc.tokenSecuredConnectorClient.SignCSR(configuration.Token.Token, csr)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to sign CSR")
	}

	credentials, err := certificates.NewCredentials(key, certResponse)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to sign CSR")
	}

	return EstablishedConnection{
		Credentials: credentials,
		DirectorURL: cc.connectorURL,
	}, nil
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
