package knativechannel

import (
	messagingV1Alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("knativechannel-controller")

const (
	controllerAgentName = "knativechannel-controller"
)

// ProvideController returns an Knative-channel controller.
func ProvideController(mgr manager.Manager) (controller.Controller, error) {

	var err error

	// Setup a new controller to Reconcile Knative Channels.
	r := &reconciler{
		recorder: mgr.GetRecorder(controllerAgentName),
		time:     util.NewDefaultCurrentTime(),
	}
	c, err := controller.New(controllerAgentName, mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		log.Error(err, "Unable to create Knative channel controller")
		return nil, err
	}

	// Watch Knative Channels
	err = c.Watch(&source.Kind{
		Type: &messagingV1Alpha1.Channel{},
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Unable to watch Knative Channel")
		return nil, err
	}

	return c, nil
}
