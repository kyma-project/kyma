package gqlschema

import "k8s.io/api/core/v1"

type Namespace struct {
	Name         string   `json:"name"`
	Applications []string `json:"applications"`
	Labels       Labels   `json:"labels"`
	Status       v1.NamespacePhase
	Pods         []Pod    `json:"pods"`
}
