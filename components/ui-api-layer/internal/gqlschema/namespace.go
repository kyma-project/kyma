package gqlschema

type Namespace struct {
	Name         string   `json:"name"`
	Applications []string `json:"applications"`
}
