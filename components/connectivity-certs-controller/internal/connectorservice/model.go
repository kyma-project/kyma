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
