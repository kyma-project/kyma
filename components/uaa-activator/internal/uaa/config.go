package uaa

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Config holds configuration for the UAA domain
type Config struct {
	ServiceInstanceParamsSecret v1beta1.SecretKeyReference
	ServiceInstance             client.ObjectKey
	ServiceBinding              client.ObjectKey
	ClusterServiceClassName     string `envconfig:"default=xsuaa"`
	ClusterServicePlanName      string `envconfig:"default=z54zhz47zdx5loz51z6z58zhvcdz59-b207b177b40ffd4b314b30635590e0ad"`
}
