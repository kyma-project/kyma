package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type namespaceConverter struct{}

func (c *namespaceConverter) AddEnvLabel(in gqlschema.Labels) {
	if _, ok := in["env"]; !ok {
		in["env"] = "true"
	}
}
