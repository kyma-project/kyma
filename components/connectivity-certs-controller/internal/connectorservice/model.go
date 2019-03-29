package connectorservice

import "github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"

type CertificatesResponse struct {
	CRTChain  string `json:"crt"`
	ClientCRT string `json:"clientCrt"`
	CaCRT     string `json:"caCrt"`
}

type InfoResponse struct {
	CsrURL          string          `json:"csrUrl"`
	Api             APIUrls         `json:"api"`
	CertificateInfo CertificateInfo `json:"certificate"`
}

type APIUrls struct {
	InfoURL         string `json:"infoUrl"`
	CertificatesURL string `json:"certificatesUrl"`
}

type CertificateInfo struct {
	Subject      string `json:"subject"`
	Extensions   string `json:"extensions"`
	KeyAlgorithm string `json:"key-algorithm"`
}

type CertificateRequest struct {
	CSR string `json:"csr"`
}

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type EstablishedConnection struct {
	Certificates      certificates.Certificates
	ManagementInfoURL string
}

type ManagementInfo struct {
	ClientIdentity ClientIdentity `json:"clientIdentity"`
	ManagementURLs ManagementURLs `json:"urls"`
}

type ClientIdentity struct {
	Application string
	Tenant      string
	Group       string
}

type ManagementURLs struct {
	EventsURL     string `json:"eventsUrl"`
	MetadataURL   string `json:"metadataUrl"`
	RenewalURL    string `json:"renewCertUrl"`
	RevocationURL string `json:"revocationCertUrl"`
}
