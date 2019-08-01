package shared

type AddonsConfiguration struct {
	Name   string
	Urls   []string
	Labels map[string]string
	Status AddonsConfigurationStatus
}

type AddonsConfigurationStatus struct {
	Phase        string
	Repositories []AddonsConfigurationRepository
}

type AddonsConfigurationRepository struct {
	Url    string
	Status string
	Addons []AddonsConfigurationAddons
}

type AddonsConfigurationAddons struct {
	Name    string
	Version string
	Status  string
	Reason  string
	Message string
}
