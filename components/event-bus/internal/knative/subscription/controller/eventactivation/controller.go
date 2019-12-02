package eventactivation

import (
	eventactivationv1alpha1 "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/applicationconnector.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("eventactivation-controller")

const (
	controllerAgentName = "eventactivation-controller"
)

// ProvideController instantiates a reconciler which reconciles EventActivations.
func ProvideController(mgr manager.Manager) error {

	var err error

	// Setup a new controller to Reconcile EventActivation.
	r := &reconciler{
		recorder: mgr.GetRecorder(controllerAgentName),
		time:     util.NewDefaultCurrentTime(),
	}
	c, err := controller.New(controllerAgentName, mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		log.Error(err, "Unable to create controller")
		return err
	}

	// Watch EventActivations.
	err = c.Watch(&source.Kind{
		Type: &eventactivationv1alpha1.EventActivation{},
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Unable to watch EventActivation")
		return err
	}

	return nil
}
