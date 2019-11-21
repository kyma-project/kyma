package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"knative.dev/pkg/injection/sharedmain"

	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/eventactivation"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/knativesubscription"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/subscription"
)

func main() {

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("entrypoint")

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
