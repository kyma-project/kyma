package process

import (
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/pkg/errors"
)

var _ Step = &PatchConnectivityValidators{}

const (
	appReleaseNameKey    = "meta.helm.sh/release-name"
	dashboardsLabelKey   = "kyma-project.io/dashboard"
	dashboardsLabelValue = "eventing"
)

type PatchConnectivityValidators struct {
	name    string
	process *Process
}

func NewPatchConnectivityValidators(p *Process) PatchConnectivityValidators {
	return PatchConnectivityValidators{
		name:    "Patch Application Connectivity Validators",
		process: p,
	}
}

func (s PatchConnectivityValidators) Do() error {
	for _, validator := range s.process.State.ConnectivityValidators.Items {
		patchedData, err := getDeploymentPatchData(validator)
		if err != nil {
			return err
		}
		_, err = s.process.Clients.Deployment.Patch(validator.Namespace, validator.Name, patchedData)
		if err != nil {
			return errors.Wrapf(err, "failed to patch validator deployment %s/%s", validator.Namespace, validator.Name)
		}
		s.process.Logger.Infof("Step: %s, patched connectivity validator: %s/%s deployment", s.ToString(), validator.Namespace, validator.Name)
	}
	return nil
}

func getNewContainerArgs(appName string) []string {
	args := []string{
		"/applicationconnectivityvalidator",
		"--proxyPort=8081",
		"--externalAPIPort=8080",
		"--tenant=",
		"--group=",
		fmt.Sprintf("--eventServicePathPrefixV1=/%s/v1/events", appName),
		fmt.Sprintf("--eventServicePathPrefixV2=/%s/v2/events", appName),
		"--eventServiceHost=eventing-event-publisher-proxy.kyma-system",
		"--eventMeshHost=eventing-event-publisher-proxy.kyma-system",
		"--eventMeshDestinationPath=/publish",
		fmt.Sprintf("--eventMeshPathPrefix=/%s/events", appName),
		fmt.Sprintf("--appRegistryPathPrefix=/%s/v1/metadata", appName),
		"--appRegistryHost=application-registry-external-api:8081",
		fmt.Sprintf("--appName=%s", appName),
		"--cacheExpirationMinutes=1",
		"--cacheCleanupMinutes=2",
	}
	return args
}

func getDeploymentPatchData(oldDeployment appsv1.Deployment) ([]byte, error) {
	appName := oldDeployment.Annotations[appReleaseNameKey]

	desiredSpec := oldDeployment.Spec.DeepCopy()
	desiredSpec.Template.Spec.Containers[0].Args = getNewContainerArgs(appName) // validator has one container only
	desiredSpec.Template.ObjectMeta.Labels = withLabels(desiredSpec.Template.ObjectMeta.Labels)
	desiredSpec.Template.ObjectMeta.Labels[dashboardsLabelKey] = dashboardsLabelValue
	targetPatch := PatchDeploymentSpec{Spec: *desiredSpec}

	containerData, err := json.Marshal(targetPatch)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "failed to marshal labels from validator deployment for app %s  patch data", appName)
	}
	return containerData, nil
}

func (s PatchConnectivityValidators) ToString() string {
	return s.name
}

func withLabels(labels Labels) Labels {
	if len(labels) == 0 {
		return make(Labels)
	}
	return labels
}
