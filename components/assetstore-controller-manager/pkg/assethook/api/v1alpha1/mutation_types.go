package v1alpha1

import "encoding/json"

type MutationRequest struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Metadata  *json.RawMessage  `json:"metadata"`
	Assets    map[string]string `json:"assets"`
}

type MutationResponse struct {
	Assets map[string]string `json:"assets"`
}
