package externalapi

type certRequest struct {
	CSR string `json:"csr"`
}

// TODO - extend response
type certResponse struct {
	CRT string `json:"crt"`
}

type infoResponse struct {
	CsrURL          string      `json:"csrUrl"`
	API             interface{} `json:"api"`
	CertificateInfo certInfo    `json:"certificate"`
}

type applicationApi struct {
	MetadataURL     string `json:"metadataUrl"`
	EventsURL       string `json:"eventsUrl"`
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
