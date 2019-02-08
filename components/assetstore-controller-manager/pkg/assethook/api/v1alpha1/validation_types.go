package v1alpha1

import "encoding/json"

type ValidationRequest struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Metadata  *json.RawMessage  `json:"metadata"`
	Assets    map[string]string `json:"assets"`
}

type ValidationResponse struct {
	Status map[string]ValidationResponseStatus `json:"status"`
}

type ValidationResponseStatus struct {
	Status  ValidationStatus `json:"status"`
	Message string           `json:"message"`
}

type ValidationStatus string

const (
	ValidationSuccess ValidationStatus = "Success"
	ValidationFailure ValidationStatus = "Failure"
)
