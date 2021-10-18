package externalapi

import (
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/clientcontext"
)

type certRequest struct {
	CSR string `json:"csr"`
}

type certResponse struct {
	CRTChain  string `json:"crt"`
	ClientCRT string `json:"clientCrt"`
	CaCRT     string `json:"caCrt"`
}

type csrInfoResponse struct {
	CsrURL          string   `json:"csrUrl"`
	API             api      `json:"api"`
	CertificateInfo certInfo `json:"certificate"`
}

type mgmtInfoReponse struct {
	ClientIdentity  interface{} `json:"clientIdentity"`
	URLs            mgmtURLs    `json:"urls"`
	CertificateInfo certInfo    `json:"certificate"`
}

type mgmtURLs struct {
	*clientcontext.RuntimeURLs
	RenewCertURL  string `json:"renewCertUrl"`
	RevokeCertURL string `json:"revokeCertUrl"`
}

type api struct {
	*clientcontext.RuntimeURLs
	InfoURL         string `json:"infoUrl"`
	CertificatesURL string `json:"certificatesUrl"`
}

type certInfo struct {
	Subject      string `json:"subject"`
	Extensions   string `json:"extensions"`
	KeyAlgorithm string `json:"key-algorithm"`
}

func toCertResponse(encodedChain certificates.EncodedCertificateChain) certResponse {
	return certResponse{
		CRTChain:  encodedChain.CertificateChain,
		ClientCRT: encodedChain.ClientCertificate,
		CaCRT:     encodedChain.CaCertificate,
	}
}
