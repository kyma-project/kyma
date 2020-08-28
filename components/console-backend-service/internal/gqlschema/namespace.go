package gqlschema

type Namespace struct {
	Name              string       `json:"name"`
	Applications      []string     `json:"applications"`
	Labels            Labels       `json:"labels"`
	Status            string       `json:"status"`
	IsSystemNamespace bool         `json:"isSystemNamespace"`
	Pods              []Pod        `json:"pods"`
	Deployments       []Deployment `json:"deployments"`
}

type NamespaceListItem struct {
	Name              string   `json:"name"`
	Applications      []string `json:"applications"`
	Labels            Labels   `json:"labels"`
	Status            string   `json:"status"`
	IsSystemNamespace bool     `json:"isSystemNamespace"`
	PodsCount         int      `json:"podsCount"`
	HealthyPodsCount  int      `json:"healthyPodsCount"`
	ApplicationsCount int      `json:"applicationsCount"`
}
