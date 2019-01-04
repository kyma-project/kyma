package api

type TokenResponse struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

type CertificateResponse struct {
	CRT string `json:"crt"`
}

type CertificateInfo struct {
	Subject      string `json:"subject"`
	Extensions   string `json:"extensions"`
	KeyAlgorithm string `json:"key-algorithm"`
}
