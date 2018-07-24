package externalapi

type certRequest struct {
	CSR string `json:"csr"`
}

type certResponse struct {
	CRT string `json:"crt"`
}

type infoResponse struct {
	SignUrl         string   `json:"csrUrl"`
	Api             api      `json:"api"`
	CertificateInfo certInfo `json:"certificate"`
}

type api struct {
	MetadataURL     string `json:"metadataUrl"`
	EventsURL       string `json:"eventsUrl"`
	CertificatesUrl string `json:"certificatesUrl"`
}

type certInfo struct {
	Subject      string `json:"subject"`
	Extensions   string `json:"extensions"`
	KeyAlgorithm string `json:"key-algorithm"`
}

type csrInfo struct {
	Country            string `json:"country"`
	Organization       string `json:"organization"`
	OrganizationalUnit string `json:"organizationalUnit"`
	Locality           string `json:"locality"`
	Province           string `json:"province"`
}
