package types

import v1 "k8s.io/api/core/v1"

type NamespaceWithAdditionalData struct {
	Namespace         *v1.Namespace
	IsSystemNamespace bool
}