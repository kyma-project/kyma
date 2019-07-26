package compass

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization"
)

// TODO: consider ConfigClient fetching client credentials from secret and Connector saving the secret
// This may imply saving DirectorURL also into the secret

//go:generate mockery -name=ConfigClient
type ConfigClient interface {
	FetchConfiguration(directorURL string, credentials certificates.Credentials) ([]synchronization.Application, error)
}

func NewConfigurationClient() ConfigClient {
	return &configClient{}
}

type configClient struct {
}

func (cc *configClient) FetchConfiguration(directorURL string, credentials certificates.Credentials) ([]synchronization.Application, error) {
	// TODO: implement

	return nil, nil
}
