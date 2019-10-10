package gqlschema

type Namespace struct {
	Name              string   `json:"name"`
	Applications      []string `json:"applications"`
	Labels            Labels   `json:"labels"`
	Status            string   `json:"status"`
	IsSystemNamespace bool     `json:"isSystemNamespace"`
	Pods              []Pod    `json:"pods"`
}
