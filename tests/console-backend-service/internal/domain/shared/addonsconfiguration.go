package shared

type AddonsConfiguration struct {
	Name   string                    `json:"name"`
	Urls   []string                  `json:"urls"`
	Labels map[string]string         `json:"labels"`
	Status AddonsConfigurationStatus `json:"status"`
}

type AddonsConfigurationStatus struct {
	Phase        string                          `json:"phase"`
	Repositories []AddonsConfigurationRepository `json:"repositories"`
}

type AddonsConfigurationRepository struct {
	Url    string                      `json:"url"`
	Status string                      `json:"status"`
	Addons []AddonsConfigurationAddons `json:"addons"`
}

type AddonsConfigurationAddons struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}
