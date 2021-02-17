package process

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/pkg/errors"
)

var _ Step = &PatchConnectivityValidators{}

const (
	appReleaseNameKey = "meta.helm.sh/release-name"
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
	var oldContainer corev1.Container
	for _, c := range oldDeployment.Spec.Template.Spec.Containers {
		oldContainer = c
	}

	desiredContainer := oldContainer.DeepCopy()
	desiredContainer.Args = getNewContainerArgs(appName)

	targetPatch := PatchDeploymentSpec{Spec: Spec{Template: Template{Spec: TemplateSpec{
		Containers: []corev1.Container{
			*desiredContainer,
		},
	}}}}

	containerData, err := json.Marshal(targetPatch)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "failed to marshal labels from validator deployment for app %s  patch data", appName)
	}
	return containerData, nil
}

func (s PatchConnectivityValidators) ToString() string {
	return s.name
}
