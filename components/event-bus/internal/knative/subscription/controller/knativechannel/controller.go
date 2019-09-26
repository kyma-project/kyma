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

// ProvideController instantiates a reconciler which reconciles Knative Channels.
func ProvideController(mgr manager.Manager) error {

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
		log.Error(err, "unable to create Knative channel controller")
		return err
	}

	// Watch Knative Channels
	err = c.Watch(&source.Kind{
		Type: &messagingV1Alpha1.Channel{},
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "unable to watch Knative Channel")
		return err
	}

	return nil
}
