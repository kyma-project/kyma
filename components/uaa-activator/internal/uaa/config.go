package uaa

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Config holds configuration for the UAA domain
type Config struct {
	ServiceInstance         client.ObjectKey
	ServiceBinding          client.ObjectKey
	DeveloperGroup          string
	DeveloperRole           string
	NamespaceAdminGroup     string
	NamespaceAdminRole      string
	IsUpgrade               bool
	ClusterServiceClassName string `envconfig:"default=xsuaa"`
	ClusterServicePlanName  string `envconfig:"default=z54zhz47zdx5loz51z6z58zhvcdz59-b207b177b40ffd4b314b30635590e0ad"`
}
