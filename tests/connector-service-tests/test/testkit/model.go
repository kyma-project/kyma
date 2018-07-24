package testkit

type TokenResponse struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

type InfoResponse struct {
	CertUrl     string   `json:"csrUrl"`
	Api         ApiInfo  `json:"api"`
	Certificate CertInfo `json:"certificate"`
}

type ApiInfo struct {
	MetadataURL     string `json:"metadataUrl"`
	EventsURL       string `json:"eventsUrl"`
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
	Crt string `json:"crt"`
}

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type Error struct {
	StatusCode    int
	ErrorResponse ErrorResponse
}
