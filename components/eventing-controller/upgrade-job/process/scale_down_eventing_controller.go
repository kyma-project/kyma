package process

import (
	"github.com/pkg/errors"
	"time"
)

var _ Step = &ScaleDownEventingController{}

// ScaleDownEventingController struct implements the interface Step
type ScaleDownEventingController struct {
	name    string
	process *Process
}

// NewScaleDownEventingController returns new instance of NewScaleDownEventingController struct
func NewScaleDownEventingController(p *Process) ScaleDownEventingController {
	return ScaleDownEventingController{
		name:    "Scale down eventing controller to zero replicas",
		process: p,
	}
}

// ToString returns step name
func (s ScaleDownEventingController) ToString() string {
	return s.name
}

// Do scales down eventing-controller deployment to zero replicas
func (s ScaleDownEventingController) Do() error {
	// Get eventing-controller deployment object
	oldDeployment, err := s.process.Clients.Deployment.Get(s.process.KymaNamespace, s.process.ControllerName)
	if err != nil {
		return err
	}

	// reduce replica count to zero
	desiredContainer := oldDeployment.DeepCopy()
	desiredContainer.Spec.Replicas = int32Ptr(0)
	_, err = s.process.Clients.Deployment.Update(s.process.KymaNamespace, desiredContainer)
	if err != nil {
		return err
	}

	// @TODO: Do we need to wait
	// Wait until pod down
	isScaledDownSuccess := false
	start := time.Now()
	for time.Since(start) < s.process.TimeoutPeriod {
		s.process.Logger.WithContext().Debug("Checking replica count of eventing-controller...")

		time.Sleep(1 * time.Second)

		controllerDeployment, err := s.process.Clients.Deployment.Get(s.process.KymaNamespace, s.process.ControllerName)
		if err != nil {
			s.process.Logger.WithContext().Error(err)
			// @TODO: should we stop or continue
			continue
		}

		if controllerDeployment.Status.Replicas == 0 {
			s.process.Logger.WithContext().Info("Eventing Controller scaled down to zero!")
			isScaledDownSuccess = true
			break
		}
	}

	if !isScaledDownSuccess {
		return errors.New("Timeout! Failed to scale down eventing-controller")
	}

	return nil
}
