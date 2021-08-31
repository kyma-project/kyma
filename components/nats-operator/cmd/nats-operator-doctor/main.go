package main

import (
	"context"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kyma-project/kyma/components/nats-operator/logger"
	"github.com/kyma-project/kyma/components/nats-operator/options"
	"github.com/kyma-project/kyma/components/nats-operator/pkg/client/natscluster"
	"github.com/kyma-project/kyma/components/nats-operator/pkg/doctor"
)

func main() {
	// setup
	log := logger.New()
	ctx := context.Background()
	opts := options.New().Parse()
	k8sConfig := config.GetConfigOrDie()
	k8sClient := kubernetes.NewForConfigOrDie(k8sConfig)
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	natsClient := natscluster.NewClient(dynamicClient)

	// start
	if err := doctor.New(k8sClient, natsClient, opts.Interval, log).Start(ctx); err != nil {
		log.Fatalf("nats-operator health-check failed with error: %v", err)
	}
}
