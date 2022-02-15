package servicebinding

import (
	"context"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceBinding struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func (sb *ServiceBinding) GetName() string {
	return sb.name
}

func New(name string, c shared.Container) *ServiceBinding {

	return &ServiceBinding{
		resCli:      resource.New(c.DynamicCli, v1beta1.SchemeGroupVersion.WithResource("servicebindings"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
	}
}

func (sb *ServiceBinding) Create(serviceInstanceName string) error {
	ac := &v1beta1.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBinding",
			APIVersion: v1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sb.name,
			Namespace: sb.namespace,
		},
		Spec: v1beta1.ServiceBindingSpec{
			InstanceRef: v1beta1.LocalObjectReference{
				Name: serviceInstanceName,
			},
		},
	}

	_, err := sb.resCli.Create(ac)
	if err != nil {
		return errors.Wrapf(err, "while creating ServiceBinding %s in namespace %s", sb.name, sb.namespace)
	}

	return err
}

func (sb *ServiceBinding) Delete() error {
	err := sb.resCli.Delete(sb.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ServiceBinding %s in namespace %s", sb.name, sb.namespace)
	}

	return nil
}

func (sb *ServiceBinding) Get() (*v1beta1.ServiceBinding, error) {
	u, err := sb.resCli.Get(sb.name)
	if err != nil {
		return &v1beta1.ServiceBinding{}, errors.Wrapf(err, "while getting ServiceBinding %s in namespace %s", sb.name, sb.namespace)
	}

	servicebinding, err := convertFromUnstructuredToServiceBinding(u)
	if err != nil {
		return &v1beta1.ServiceBinding{}, err
	}

	return &servicebinding, nil
}

func (sb *ServiceBinding) LogResource() error {
	serviceBinding, err := sb.Get()
	if err != nil {
		return err
	}
	out, err := helpers.PrettyMarshall(serviceBinding)
	if err != nil {
		return err
	}

	sb.log.Infof("Service Binding resource: %s", out)
	return nil
}

func (sb *ServiceBinding) WaitForStatusRunning() error {
	servicebinding, err := sb.Get()
	if err != nil {
		return err
	}

	// we need to ensure that status is ready first, because otherwise we would not Get any events in watchtools.Until
	if sb.isReadyPhase(*servicebinding) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), sb.waitTimeout)
	defer cancel()
	condition := sb.isServiceBindingReady()
	return resource.WaitUntilConditionSatisfied(ctx, sb.resCli.ResCli, condition)
}

func (sb *ServiceBinding) isServiceBindingReady() func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type != watch.Modified {
			return false, nil
		}
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if u.GetName() != sb.name {
			return false, nil
		}

		servicebinding, err := convertFromUnstructuredToServiceBinding(u)
		if err != nil {
			return false, err
		}

		return sb.isReadyPhase(servicebinding), nil
	}
}

func (sb *ServiceBinding) isReadyPhase(servicebinding v1beta1.ServiceBinding) bool {
	if len(servicebinding.Status.Conditions) == 0 {
		shared.LogReadiness(false, sb.verbose, sb.name, sb.log, servicebinding)
		return false
	}

	ready := false
	for _, condition := range servicebinding.Status.Conditions {
		if condition.Type == v1beta1.ServiceBindingConditionReady && condition.Status == v1beta1.ConditionTrue {
			ready = true
		}
	}

	shared.LogReadiness(ready, sb.verbose, sb.name, sb.log, servicebinding)

	return ready
}

func convertFromUnstructuredToServiceBinding(u *unstructured.Unstructured) (v1beta1.ServiceBinding, error) {
	servicebinding := v1beta1.ServiceBinding{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &servicebinding)
	return servicebinding, err
}
