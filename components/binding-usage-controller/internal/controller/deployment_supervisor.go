package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/pretty"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsV1beta2 "k8s.io/api/apps/v1beta2"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	informerV1beta2 "k8s.io/client-go/informers/apps/v1beta2"
	clientAppsV1beta2 "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
	listerV1beta2 "k8s.io/client-go/listers/apps/v1beta2"
	"k8s.io/client-go/tools/cache"
)

// DeploymentSupervisor validates if Deployment can be modified by given ServiceBindingUsage. If yes
// it can ensure that labels are present or deleted on Deployment
type DeploymentSupervisor struct {
	deploymentLister   listerV1beta2.DeploymentLister
	deploymenHasSynced cache.InformerSynced
	appsClient         clientAppsV1beta2.AppsV1beta2Interface
	log                logrus.FieldLogger

	tracer usageBindingAnnotationTracer
}

// NewDeploymentSupervisor creates a new DeploymentSupervisor.
func NewDeploymentSupervisor(deploymentInformer informerV1beta2.DeploymentInformer, appsClient clientAppsV1beta2.AppsV1beta2Interface, log logrus.FieldLogger) *DeploymentSupervisor {
	return &DeploymentSupervisor{
		deploymentLister:   deploymentInformer.Lister(),
		deploymenHasSynced: deploymentInformer.Informer().HasSynced,
		appsClient:         appsClient,
		log:                log.WithField("service", "controller:deployment-supervisor"),

		tracer: &usageAnnotationTracer{},
	}
}

// EnsureLabelsCreated ensures that given labels are added to Deployment
func (m *DeploymentSupervisor) EnsureLabelsCreated(deployNS, deployName, usageName string, labels map[string]string) error {
	cacheDeploy, err := m.deploymentLister.Deployments(deployNS).Get(deployName)
	if err != nil {
		return errors.Wrap(err, "while getting Deployment")
	}

	oldDeploy := cacheDeploy.DeepCopy()
	newDeploy := cacheDeploy.DeepCopy()

	// check labels conflicts
	if conflictsKeys, found := detectLabelsConflicts(m.convertToLabeler(newDeploy), labels); found {
		err := fmt.Errorf("found conflicts in %s under 'Spec.Template.Labels' field: %s keys already exists [override forbidden]",
			pretty.DeploymentName(newDeploy), strings.Join(conflictsKeys, ","))
		return err
	}

	// apply new labels
	newDeploy.Spec.Template.Labels = EnsureMapIsInitiated(newDeploy.Spec.Template.Labels)
	for k, v := range labels {
		newDeploy.Spec.Template.Labels[k] = v
	}

	if err := m.tracer.SetAnnotationAboutBindingUsage(&newDeploy.ObjectMeta, usageName, labels); err != nil {
		return errors.Wrap(err, "while setting annotation tracing data")
	}

	if err := m.patchDeployment(oldDeploy, newDeploy); err != nil {
		return errors.Wrap(err, "while patching deployment")
	}

	return nil
}

// EnsureLabelsDeleted ensures that given labels are deleted on Deployment
func (m *DeploymentSupervisor) EnsureLabelsDeleted(deployNs, deployName, usageName string) error {
	cacheDeploy, err := m.deploymentLister.Deployments(deployNs).Get(deployName)
	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		return nil
	default:
		return errors.Wrap(err, "while getting Deployment")
	}

	oldDeploy := cacheDeploy.DeepCopy()
	newDeploy := cacheDeploy.DeepCopy()

	// remove old labels
	err = m.ensureOldLabelsAreRemovedOnNewDeploy(oldDeploy, newDeploy, usageName)
	if err != nil {
		return errors.Wrap(err, "while trying to delete old labels")
	}

	// remove annotations
	err = m.tracer.DeleteAnnotationAboutBindingUsage(&newDeploy.ObjectMeta, usageName)
	if err != nil {
		return errors.Wrap(err, "while deleting annotation tracing data")
	}

	if err := m.patchDeployment(oldDeploy, newDeploy); err != nil {
		return errors.Wrap(err, "while patching deployment")
	}

	return nil
}

// GetInjectedLabels returns labels applied on given Deployment by usage controller
func (m *DeploymentSupervisor) GetInjectedLabels(deployNS, deployName, usageName string) (map[string]string, error) {
	deployments, err := m.deploymentLister.Deployments(deployNS).Get(deployName)
	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		return nil, NewNotFoundError(err.Error())
	default:
		return nil, errors.Wrap(err, "while listing Deployments")
	}

	labels, err := m.tracer.GetInjectedLabels(deployments.ObjectMeta, usageName)
	if err != nil {
		return nil, errors.Wrap(err, "while getting injected labels keys")
	}

	return labels, nil
}

// HasSynced returns true if the shared informer's store has synced
func (m *DeploymentSupervisor) HasSynced() bool {
	return m.deploymenHasSynced()
}

func (m *DeploymentSupervisor) ensureOldLabelsAreRemovedOnNewDeploy(oldDeploy, newDeploy *appsV1beta2.Deployment, usageName string) error {
	labels, err := m.tracer.GetInjectedLabels(oldDeploy.ObjectMeta, usageName)
	if err != nil {
		return errors.Wrap(err, "while getting injected labels")
	}
	if len(labels) == 0 {
		return nil
	}

	if newDeploy.Spec.Template.Labels == nil {
		m.log.Warningf("missing labels from previous modification")
		return nil
	}

	for lk := range labels {
		delete(newDeploy.Spec.Template.Labels, lk)
	}
	return nil
}

func (m *DeploymentSupervisor) patchDeployment(oldDeploy, newDeploy *appsV1beta2.Deployment) error {
	oldData, err := json.Marshal(oldDeploy)
	if err != nil {
		return errors.Wrap(err, "while marshaling old deployment")
	}

	newData, err := json.Marshal(newDeploy)
	if err != nil {
		return errors.Wrap(err, "while marshaling new deployment")
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, appsV1beta2.Deployment{})
	if err != nil {
		return errors.Wrap(err, "while creating marge patch")
	}

	_, err = m.appsClient.Deployments(newDeploy.Namespace).Patch(newDeploy.Name, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return errors.Wrapf(err, "while patching %s", pretty.DeploymentName(newDeploy))
	}

	return nil
}

func (m *DeploymentSupervisor) convertToLabeler(deploy *appsV1beta2.Deployment) *labelledDeployment {
	return (*labelledDeployment)(deploy)
}

type labelledDeployment appsV1beta2.Deployment

func (l *labelledDeployment) Labels() map[string]string {
	return l.Spec.Template.Labels
}
