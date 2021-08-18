package process

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
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
	if s.process.State.Is124Cluster {
		// Scale down
		return s.ScaleDownK8sDeployment(s.process.KymaNamespace, s.process.ControllerName)
	}

	s.process.Logger.WithContext().Info(fmt.Sprintf("Skipping step: %s, because it is not a v1.24.x cluster", s.ToString()))
	return nil
}

// ScaleDownK8sDeployment scales down k8s deployment to zero replicas
func (s ScaleDownEventingController) ScaleDownK8sDeployment(namespace, name string) error {

	// Create patch object to reduce replica count to zero
	targetPatch := PatchDeploymentSpec{
		Spec: Spec{
			Replicas: int32Ptr(0),
		},
	}
	containerData, err := json.Marshal(targetPatch)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal deployment patch for %s", s.process.ControllerName)
	}

	// Patch the eventing controller deployment
	_, err = s.process.Clients.Deployment.Patch(namespace, name, containerData)
	if err != nil {
		return err
	}

	// Wait until pod down
	isScaledDownSuccess := false
	start := time.Now()
	for time.Since(start) < s.process.TimeoutPeriod {
		s.process.Logger.WithContext().Infof("Checking replica count of deployment: %s", name)

		time.Sleep(5 * time.Second)

		deployment, err := s.process.Clients.Deployment.Get(namespace, name)
		if err != nil {
			s.process.Logger.WithContext().Error(err)
			continue
		}

		if deployment.Status.Replicas == 0 {
			s.process.Logger.WithContext().Info(fmt.Sprintf("Deployment: %s scaled down to zero!", name))
			isScaledDownSuccess = true
			break
		}
	}

	if !isScaledDownSuccess {
		return errors.New(fmt.Sprintf("Timeout! Failed to scale down deployment: %s", name))
	}

	return nil
}
