package utils

type Subject struct {
	CommonName         string
	Country            string
	Organization       string
	OrganizationalUnit string
	Locality           string
	Province           string
}

type InfoResponse struct {
	CertUrl     string   `json:"csrUrl"`
	Api         ApiInfo  `json:"api"`
	Certificate CertInfo `json:"certificate"`
}

type RuntimeURLs struct {
	MetadataUrl string `json:"metadataUrl"`
	EventsUrl   string `json:"eventsUrl"`
}

type ApiInfo struct {
	*RuntimeURLs
	ManagementInfoURL string `json:"infoUrl"`
	CertificatesUrl   string `json:"certificatesUrl"`
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
