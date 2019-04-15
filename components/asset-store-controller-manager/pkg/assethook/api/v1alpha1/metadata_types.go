package v1alpha1

import "encoding/json"

// ResultError stores error data
type MetadataResultError struct {
	FilePath string `json:"filePath,omitempty"`
	Message  string `json:"message,omitempty"`
}

// ResultSuccess stores success data
type MetadataResultSuccess struct {
	FilePath string           `json:"filePath,omitempty"`
	Metadata *json.RawMessage `json:"metadata,omitempty"`
}

type MetadataResponse struct {
	Data   []MetadataResultSuccess `json:"data,omitempty"`
	Errors []MetadataResultError   `json:"errors,omitempty"`
}
