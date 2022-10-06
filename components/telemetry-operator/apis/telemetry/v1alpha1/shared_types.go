package v1alpha1

import "k8s.io/apimachinery/pkg/types"

type ValueType struct {
	Value     string           `json:"value,omitempty"`
	ValueFrom *ValueFromSource `json:"valueFrom,omitempty"`
}

func (v *ValueType) IsDefined() bool {
	if v.Value != "" {
		return true
	}

	return v.ValueFrom != nil && v.ValueFrom.IsSecretKeyRef()
}

type ValueFromSource struct {
	SecretKeyRef *SecretKeyRef `json:"secretKeyRef,omitempty"`
}

func (v *ValueFromSource) IsSecretKeyRef() bool {
	return v.SecretKeyRef != nil && v.SecretKeyRef.Name != "" && v.SecretKeyRef.Key != ""
}

type SecretKeyRef struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Key       string `json:"key,omitempty"`
}

func (skr *SecretKeyRef) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: skr.Name, Namespace: skr.Namespace}
}
