package handlers

import (
	"log"
	"net/http"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingv1alpha1client "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
)

// ReadinessProbeHandler of the Knative PublishApplication
func ReadinessProbeHandler(msgClientInf messagingv1alpha1client.MessagingV1alpha1Interface) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Namespace could be anything but hardcoded to default as it is a readiness check call to Kube
		// API server and the output is ignored
		_, err := msgClientInf.Channels("default").List(v1.ListOptions{})
		if err != nil {
			log.Printf("Error in ReadinessProbeHandler: %v", err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
