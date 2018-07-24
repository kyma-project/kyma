package controller

import (
	"fmt"
	"strings"

	kubelessTypes "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubelessClient "github.com/kubeless/kubeless/pkg/client/clientset/versioned/typed/kubeless/v1beta1"
	kubelessInformer "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
	kubelessLister "github.com/kubeless/kubeless/pkg/client/listers/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/pretty"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
)

// KubelessFunctionSupervisor validates if Function can be modified by given ServiceBindingUsage. If yes
// it can ensure that labels are present or deleted on Deployment for given Function
type KubelessFunctionSupervisor struct {
	functionLister       kubelessLister.FunctionLister
	functionListerSynced cache.InformerSynced
	kubelessClient       kubelessClient.KubelessV1beta1Interface
	log                  logrus.FieldLogger
	tracer               usageBindingAnnotationTracer
}

// NewKubelessFunctionSupervisor creates a new KubelessFunctionSupervisor.
func NewKubelessFunctionSupervisor(fnInformer kubelessInformer.FunctionInformer, kubelessClient kubelessClient.KubelessV1beta1Interface, log logrus.FieldLogger) *KubelessFunctionSupervisor {
	return &KubelessFunctionSupervisor{
		functionLister:       fnInformer.Lister(),
		functionListerSynced: fnInformer.Informer().HasSynced,
		kubelessClient:       kubelessClient,
		log:                  log.WithField("service", "controller:kubeless-function-supervisor"),

		tracer: &usageAnnotationTracer{},
	}
}

// EnsureLabelsCreated ensures that given labels are added to Deployment for given Function
func (m *KubelessFunctionSupervisor) EnsureLabelsCreated(fnNs, fnName, usageName string, labels map[string]string) error {
	cacheFn, err := m.functionLister.Functions(fnNs).Get(fnName)
	if err != nil {
		return errors.Wrap(err, "while getting Function")
	}

	newFn := cacheFn.DeepCopy()

	// check labels conflicts
	if conflictsKeys, found := detectLabelsConflicts(m.convertToLabeler(newFn), labels); found {
		err := fmt.Errorf("found conflicts in %s under 'Spec.Deployment.Spec.Template.Labels' field: %s keys already exists [override forbidden]",
			pretty.FunctionName(newFn), strings.Join(conflictsKeys, ","))
		return err
	}

	// apply new labels
	newFn.Spec.Deployment.Spec.Template.Labels = EnsureMapIsInitiated(newFn.Spec.Deployment.Spec.Template.Labels)
	for k, v := range labels {
		newFn.Spec.Deployment.Spec.Template.Labels[k] = v
	}

	if err := m.tracer.SetAnnotationAboutBindingUsage(&newFn.ObjectMeta, usageName, labels); err != nil {
		return errors.Wrap(err, "while setting annotation tracing data")
	}

	if err := m.updateFunction(newFn); err != nil {
		return errors.Wrap(err, "while patching Function")
	}

	return nil
}

// EnsureLabelsDeleted ensures that given labels are deleted on Deployment for given Function
func (m *KubelessFunctionSupervisor) EnsureLabelsDeleted(fnNs, fnName, usageName string) error {
	cacheFn, err := m.functionLister.Functions(fnNs).Get(fnName)
	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		return nil
	default:
		return errors.Wrap(err, "while getting Function")
	}

	oldFn := cacheFn.DeepCopy()
	newFn := cacheFn.DeepCopy()

	// remove old labels
	err = m.ensureOldLabelsAreRemovedOnNewFunction(oldFn, newFn, usageName)
	if err != nil {
		return errors.Wrap(err, "while trying to delete old labels")
	}

	// remove annotations
	err = m.tracer.DeleteAnnotationAboutBindingUsage(&newFn.ObjectMeta, usageName)
	if err != nil {
		return errors.Wrap(err, "while deleting annotation tracing data")
	}

	if err := m.updateFunction(newFn); err != nil {
		return errors.Wrap(err, "while patching Function")
	}

	return nil
}

// GetInjectedLabels returns labels applied on given Function
func (m *KubelessFunctionSupervisor) GetInjectedLabels(fnNS, fnName, usageName string) (map[string]string, error) {
	fn, err := m.functionLister.Functions(fnNS).Get(fnName)
	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		return nil, NewNotFoundError(err.Error())
	default:
		return nil, errors.Wrap(err, "while listing Deployments")
	}

	labels, err := m.tracer.GetInjectedLabels(fn.ObjectMeta, usageName)
	if err != nil {
		return nil, errors.Wrap(err, "while getting injected labels keys")
	}

	return labels, nil
}

// HasSynced returns true if the shared informer's store has synced
func (m *KubelessFunctionSupervisor) HasSynced() bool {
	return m.functionListerSynced()
}

func (m *KubelessFunctionSupervisor) ensureOldLabelsAreRemovedOnNewFunction(oldFn, newFn *kubelessTypes.Function, usageName string) error {
	labels, err := m.tracer.GetInjectedLabels(oldFn.ObjectMeta, usageName)
	if err != nil {
		return errors.Wrap(err, "while getting injected labels")
	}
	if len(labels) == 0 {
		return nil
	}

	if newFn.Spec.Deployment.Spec.Template.Labels == nil {
		m.log.Warningf("missing labels from previous modification")
		return nil
	}

	for lk := range labels {
		delete(newFn.Spec.Deployment.Spec.Template.Labels, lk)
	}

	return nil
}

// CreateTwoWayMergePatch is not supported on custom resources
func (m *KubelessFunctionSupervisor) updateFunction(newFn *kubelessTypes.Function) error {
	_, err := m.kubelessClient.Functions(newFn.Namespace).Update(newFn)
	if err != nil && !apiErrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "while updating %s", pretty.FunctionName(newFn))
	}

	return nil
}

func (m *KubelessFunctionSupervisor) convertToLabeler(fn *kubelessTypes.Function) *labelledFunction {
	return (*labelledFunction)(fn)
}

type labelledFunction kubelessTypes.Function

func (l *labelledFunction) Labels() map[string]string {
	return l.Spec.Deployment.Spec.Template.Labels
}
