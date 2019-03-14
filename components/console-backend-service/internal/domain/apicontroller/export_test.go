package apicontroller

import "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned/fake"

func (r *PluggableResolver) SetFakeClient() {
	r.cfg.client = fake.NewSimpleClientset()
}
