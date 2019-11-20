package handlers

import (
	"net/http"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/knative/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
)

// ReadinessProbeHandler of the Knative PublishApplication
func ReadinessProbeHandler(evClient eventingv1alpha1.EventingV1alpha1Interface) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Namespace could be anything but hardcoded to default as it is a readiness check call to Kube
		// API server and the output is ignored
		_, err := evClient.Channels("default").List(v1.ListOptions{})
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
