package externalapi

import "github.com/kyma-project/kyma/components/connector-service/internal/api"

type CertificateRequest struct {
	CSR         string      `json:"csr"`
	KymaCluster KymaCluster `json:"kymaCluster"`
}

type InfoResponse struct {
	SignUrl         string              `json:"csrUrl"`
	CertificateInfo api.CertificateInfo `json:"certificate"`
}

type KymaCluster struct {
	AppRegistryUrl string `json:"appRegistryUrl"`
	EventsUrl      string `json:"eventsUrl"`
}
