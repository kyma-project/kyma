package ybind

// BindYAML represents a yBundle plan bind.yaml structure
type BindYAML struct {
	Credential     []CredentialVar        `json:"credential"`
	CredentialFrom []CredentialFromSource `json:"credentialFrom"`
}

// CredentialVar represents an credential variable.
type CredentialVar struct {
	// Required
	Name string `json:"name"`
	// Optional: no more than one of the following may be specified.
	Value string `json:"value"`
	// Optional: Specifies a source the value of this var should come from.
	ValueFrom *CredentialVarSource `json:"valueFrom,omitempty"`
}

// CredentialVarSource represents a source for the value of an CredentialVar.
// ONLY ONE of its fields may be set.
type CredentialVarSource struct {
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *KeySelector `json:"configMapKeyRef,omitempty"`
	// Selects a key of a Secret in the helm release namespace.
	// +optional
	SecretKeyRef *KeySelector `json:"secretKeyRef,omitempty"`
	// Selects a field from a Service.
	// +optional
	ServiceRef *JSONPathSelector `json:"serviceRef,omitempty"`
}

// KeySelector selects a key of a k8s resource.
type KeySelector struct {
	NameSelector `json:",inline"`
	// The key of the resource to select from.
	Key string `json:"key"`
}

// JSONPathSelector select a field of a k8s resource by defining JSONPath
type JSONPathSelector struct {
	NameSelector `json:",inline"`
	// JSONPath template for extracting given value from resource.
	JSONPath string `json:"jsonpath"`
}

// CredentialFromSource represents list of sources to populate credentials variables.
type CredentialFromSource struct {
	// The ConfigMap to select from
	// + optional
	ConfigMapRef *NameSelector `json:"configMapRef,omitempty"`
	// The Secret to select from
	// +optional
	SecretRef *NameSelector `json:"secretRef,omitempty"`
}

// NameSelector selects by the name of k8s resource.
type NameSelector struct {
	// Name the name of k8s resource.
	Name string `json:"name"`
}

// RenderedBindYAML is used to represent already rendered YAML for binding.
type RenderedBindYAML []byte
