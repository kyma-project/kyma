package compassconnection

import (
	"crypto/x509/pkix"
	"strings"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
)

type EstablishedConnection struct {
	Credentials    certificates.Credentials
	ManagementInfo v1alpha1.ManagementInfo
}

const (
	ConnectorTokenHeader = "Connector-Token"
)

//go:generate mockery -name=Connector
type Connector interface {
	EstablishConnection(connectorURL, token string) (EstablishedConnection, error)
	MaintainConnection(credentials certificates.ClientCredentials, connectorURL string, renewCert bool) (*certificates.Credentials, v1alpha1.ManagementInfo, error)
}

func NewCompassConnector(
	csrProvider certificates.CSRProvider,
	clientsProvider compass.ClientsProvider,
	insecureConnectorCommunication bool,
) Connector {
	return &compassConnector{
		csrProvider:                    csrProvider,
		clientsProvider:                clientsProvider,
		insecureConnectorCommunication: insecureConnectorCommunication,
	}
}

type compassConnector struct {
	csrProvider                    certificates.CSRProvider
	clientsProvider                compass.ClientsProvider
	insecureConnectorCommunication bool
}

func (cc *compassConnector) EstablishConnection(connectorURL, token string) (EstablishedConnection, error) {
	if connectorURL == "" {
		return EstablishedConnection{}, errors.New("Failed to establish connection. Connector URL is empty")
	}

	tokenSecuredConnectorClient, err := cc.clientsProvider.GetConnectorClient(connectorURL)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to prepare Connector Token-secured client")
	}

	configuration, err := tokenSecuredConnectorClient.Configuration(connectorTokenHeader(token))
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to fetch configuration")
	}

	subject := parseSubject(configuration.CertificateSigningRequestInfo.Subject)
	csr, key, err := cc.csrProvider.CreateCSR(subject)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to generate CSR")
	}

	certResponse, err := tokenSecuredConnectorClient.SignCSR(csr, connectorTokenHeader(token))
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to sign CSR")
	}

	credentials, err := certificates.NewCredentials(key, certResponse)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to parse certification response to credentials")
	}

	return EstablishedConnection{
		Credentials:    credentials,
		ManagementInfo: toManagementInfo(configuration.ManagementPlaneInfo),
	}, nil
}

func (cc *compassConnector) MaintainConnection(credentials certificates.ClientCredentials, connectorURL string, renewCert bool) (*certificates.Credentials, v1alpha1.ManagementInfo, error) {
	certSecuredClient, err := cc.clientsProvider.GetConnectorCertSecuredClient(credentials, connectorURL)
	if err != nil {
		return nil, v1alpha1.ManagementInfo{}, errors.Wrap(err, "Failed to prepare Certificate-secured Connector client while checking connection")
	}

	configuration, err := certSecuredClient.Configuration(nil)
	if err != nil {
		return nil, v1alpha1.ManagementInfo{}, errors.Wrap(err, "Failed to query Connection Configuration while checking connection")
	}

	if !renewCert {
		return nil, toManagementInfo(configuration.ManagementPlaneInfo), nil
	}

	subject := parseSubject(configuration.CertificateSigningRequestInfo.Subject)
	csr, key, err := cc.csrProvider.CreateCSR(subject)
	if err != nil {
		return nil, v1alpha1.ManagementInfo{}, errors.Wrap(err, "Failed to create CSR while renewing connection")
	}

	certResponse, err := certSecuredClient.SignCSR(csr, nil)
	if err != nil {
		return nil, v1alpha1.ManagementInfo{}, errors.Wrap(err, "Failed to sign CSR while renewing connection")
	}

	renewedCredentials, err := certificates.NewCredentials(key, certResponse)
	if err != nil {
		return nil, v1alpha1.ManagementInfo{}, errors.Wrap(err, "Failed to parse certification response to credentials while renewing connection")
	}

	return &renewedCredentials, toManagementInfo(configuration.ManagementPlaneInfo), nil
}

func toManagementInfo(configInfo *gqlschema.ManagementPlaneInfo) v1alpha1.ManagementInfo {
	if configInfo == nil {
		return v1alpha1.ManagementInfo{}
	}

	var directorURL = ""
	if configInfo.DirectorURL != nil {
		directorURL = *configInfo.DirectorURL
	}
	var certSecuredConnectorURL = ""
	if configInfo.CertificateSecuredConnectorURL != nil {
		certSecuredConnectorURL = *configInfo.CertificateSecuredConnectorURL
	}

	return v1alpha1.ManagementInfo{
		DirectorURL:  directorURL,
		ConnectorURL: certSecuredConnectorURL,
	}
}

func connectorTokenHeader(token string) map[string][]string {
	return map[string][]string{
		ConnectorTokenHeader: {token},
	}
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
