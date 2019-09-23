package apicontroller

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
)

func (r *PluggableResolver) SetFakeClient() {
	scheme := runtime.NewScheme()
	r.cfg.client = fake.NewSimpleDynamicClient(scheme)
}
