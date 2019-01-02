package externalapi

import "github.com/kyma-project/kyma/components/connector-service/internal/api"

type CertificateRequest struct {
	CSR         string      `json:"csr"`
	Application Application `json:"application"`
}

type Application struct {
	DisplayName string            `json:"displayName"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

type InfoResponse struct {
	SignUrl         string              `json:"csrUrl"`
	Api             Api                 `json:"api"`
	CertificateInfo api.CertificateInfo `json:"certificate"`
}

type Api struct {
	MetadataURL     string `json:"metadataUrl"`
	EventsURL       string `json:"eventsUrl"`
	CertificatesUrl string `json:"certificatesUrl"`
}
