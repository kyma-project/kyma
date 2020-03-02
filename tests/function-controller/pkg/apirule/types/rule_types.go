// Copied from:
//https://github.com/ory/oathkeeper-maester/blob/master/api/v1alpha1/rule_types.go
package types

import "k8s.io/apimachinery/pkg/runtime"

// Authenticator represents a handler that authenticates provided credentials.
type Authenticator struct {
	*Handler `json:",inline"`
}

// Authorizer represents a handler that authorizes the subject ("user") from the previously validated credentials making the request.
type Authorizer struct {
	*Handler `json:",inline"`
}

// Mutator represents a handler that transforms the HTTP request before forwarding it.
type Mutator struct {
	*Handler `json:",inline"`
}

// Handler represents an Oathkeeper routine that operates on incoming requests. It is used to either validate a request (Authenticator, Authorizer) or modify it (Mutator).
type Handler struct {
	// Name is the name of a handler
	Name string `json:"handler"`
	// Config configures the handler. Configuration keys vary per handler.
	// +kubebuilder:validation:Type=object
	Config *runtime.RawExtension `json:"config,omitempty"`
}