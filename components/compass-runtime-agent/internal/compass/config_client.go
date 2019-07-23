package compass

type ConfigClient interface {
	FetchConfiguration() ([]Application, error)
}

func NewConfigurationClient() ConfigClient {
	return &configClient{}
}

type configClient struct {
}

func (cc *configClient) FetchConfiguration() ([]Application, error) {
	return nil, nil
}
