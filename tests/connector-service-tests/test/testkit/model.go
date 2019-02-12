package testkit

import "crypto/x509"

type TokenResponse struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

type InfoResponse struct {
	CertUrl     string   `json:"csrUrl"`
	Api         ApiInfo  `json:"api"`
	Certificate CertInfo `json:"certificate"`
}

type ManagementInfoResponse struct {
	MetadataUrl string `json:"metadataUrl"`
	EventsUrl string `json:"eventsUrl"`
	RenewCertUrl string `json:"renewCertUrl"`
}

type ApiInfo struct {
	MetadataURL     string `json:"metadataUrl"`
	EventsURL       string `json:"eventsUrl"`
	GetInfoURL      string `json:"infoUrl"`
	CertificatesUrl string `json:"certificatesUrl"`
}

type CertInfo struct {
	Subject      string `json:"subject"`
	Extensions   string `json:"extensions"`
	KeyAlgorithm string `json:"key-algorithm"`
}

type CsrRequest struct {
	Csr string `json:"csr"`
}

type CrtResponse struct {
	CRTChain  string `json:"crt"`
	ClientCRT string `json:"clientCrt"`
	CaCRT     string `json:"caCrt"`
}

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type Error struct {
	StatusCode    int
	ErrorResponse ErrorResponse
}

type DecodedCrtResponse struct {
	CRTChain  []*x509.Certificate
	ClientCRT *x509.Certificate
	CaCRT     *x509.Certificate
}
