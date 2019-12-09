package main

import (
	"knative.dev/pkg/injection/sharedmain"

	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/eventactivation"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/knativesubscription"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/subscription"
)

func main() {
	sharedmain.Main("eventbus_controller",
		eventactivation.NewController,
		subscription.NewController,
		knativesubscription.NewController)
}
