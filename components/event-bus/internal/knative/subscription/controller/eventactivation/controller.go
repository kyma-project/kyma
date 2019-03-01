package eventactivation

import (
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	eventactivationv1alpha1 "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/applicationconnector.kyma-project.io/v1alpha1"
)

var log = logf.Log.WithName("ea-controller")

const (
	controllerAgentName = "ea-controller"
)

// ProvideController returns an EventActivation controller.
func ProvideController(mgr manager.Manager) (controller.Controller, error) {

	var err error

	// Setup a new controller to Reconcile EventActivation.
	r := &reconciler{
		recorder: mgr.GetRecorder(controllerAgentName),
	}
	c, err := controller.New(controllerAgentName, mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		log.Error(err, "Unable to create controller")
		return nil, err
	}

	// Watch EventActivations.
	err = c.Watch(&source.Kind{
		Type: &eventactivationv1alpha1.EventActivation{},
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Unable to watch EventActivation")
		return nil, err
	}



	return c, nil
}
