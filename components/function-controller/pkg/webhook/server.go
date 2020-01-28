package webhook

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

const (
	certDir = "/tmp/k8s-webhook-server/serving-certs"
	port    = 9876
)

// Add adds itself to the manager
func Add(mgr manager.Manager) {
	srv := mgr.GetWebhookServer()
	srv.CertDir = certDir
	srv.Port = port
	srv.Register("/"+webhookEndpoint, &webhook.Admission{Handler: &FunctionCreateHandler{}})
}
