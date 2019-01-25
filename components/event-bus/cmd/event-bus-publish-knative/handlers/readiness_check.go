package handlers

import (
	"net/http"

	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

func ReadinessProbeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := knative.GetKnativeLib(); err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
