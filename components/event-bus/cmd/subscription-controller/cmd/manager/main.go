package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"

	"knative.dev/pkg/injection/sharedmain"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/eventactivation"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/knativesubscription"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/subscription"
)

func main() {
	log.SetLogger(log.ZapLogger(false))
	log := log.Log.WithName("entrypoint")


	sharedmain.Main("eventbus_controller",
		eventactivation.NewController,
		subscription.NewController,
		knativesubscription.NewController)

	// Setup Monitoring Service
	metricsServeMux := http.NewServeMux()
	metricsServeMux.Handle("/metrics", promhttp.Handler())
	metricsServer := http.Server{
		Addr:    fmt.Sprintf(":%d", 9090),
		Handler: metricsServeMux,
	}
	log.Info("HTTP metrics server starting on %v", 9090)
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil {
			log.Error(err, "HTTP metrics server failed with error|||")
		}
	}()
}

var statusLive, statusReady common.StatusReady
