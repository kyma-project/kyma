package externalapi

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
)

type certRequest struct {
	CSR string `json:"csr"`
}

type certResponse struct {
	CRTChain  string `json:"crt"`
	ClientCRT string `json:"clientCrt"`
	CaCRT     string `json:"caCrt"`
}

type infoResponse struct {
	CsrURL          string      `json:"csrUrl"`
	API             interface{} `json:"api"`
	CertificateInfo certInfo    `json:"certificate"`
}

type api struct {
	*clientcontext.RuntimeURLs
	InfoURL         string `json:"infoUrl"`
	CertificatesURL string `json:"certificatesUrl"`
}

type runtimeApi struct {
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
