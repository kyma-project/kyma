package servicebindingusage

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebindingusage/types/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceBindingUsage struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
	usageKind   string
}

func New(name, usageKind string, c shared.Container) *ServiceBindingUsage {
	return &ServiceBindingUsage{
		resCli:      resource.New(c.DynamicCli, v1alpha1.SchemeGroupVersion.WithResource("servicebindingusages"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		usageKind:   usageKind,
	}
}

func (sbu *ServiceBindingUsage) GetName() string {
	return sbu.name
}

func (sbu *ServiceBindingUsage) Create(serviceBindingName, fnKsvcName, envPrefix string) error {
	servicebindingusage := &v1alpha1.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sbu.name,
			Namespace: sbu.namespace,
		},
		Spec: v1alpha1.ServiceBindingUsageSpec{
			ServiceBindingRef: v1alpha1.LocalReferenceByName{
				Name: serviceBindingName,
			},
			UsedBy: v1alpha1.LocalReferenceByKindAndName{
				Name: fnKsvcName,
				Kind: sbu.usageKind,
			},
			Parameters: &v1alpha1.Parameters{
				EnvPrefix: &v1alpha1.EnvPrefix{Name: envPrefix},
			},
		},
	}

	_, err := sbu.resCli.Create(servicebindingusage)
	if err != nil {
		return errors.Wrapf(err, "while creating ServiceBindingUsage %s in namespace %s", sbu.name, sbu.namespace)
	}

	return err
}

func (sbu *ServiceBindingUsage) Delete() error {
	err := sbu.resCli.Delete(sbu.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ServiceBindingUsage %s in namespace %s", sbu.name, sbu.namespace)
	}

	return nil
}

func (sbu *ServiceBindingUsage) Get() (*v1alpha1.ServiceBindingUsage, error) {
	u, err := sbu.resCli.Get(sbu.name)
	if err != nil {
		return &v1alpha1.ServiceBindingUsage{}, errors.Wrapf(err, "while getting ServiceBindingUsage %s in namespace %s", sbu.name, sbu.namespace)
	}

	serviceBindingUsage, err := convertFromUnstructuredToServiceBindingUsage(u)
	if err != nil {
		return &v1alpha1.ServiceBindingUsage{}, err
	}

	return &serviceBindingUsage, nil
}

func (sbu *ServiceBindingUsage) LogResource() error {
	serviceBindingUsage, err := sbu.Get()
	if err != nil {
		return err
	}

	out, err := helpers.PrettyMarshall(serviceBindingUsage)
	if err != nil {
		return err
	}

	sbu.log.Infof("Service Binding Usage resource: %s", out)
	return nil
}

func (sbu *ServiceBindingUsage) WaitForStatusRunning() error {
	servicebinding, err := sbu.Get()
	if err != nil {
		return err
	}

	// we need to ensure that status is ready first, because otherwise we would not Get any events in watchtools.Until
	if sbu.isReadyPhase(*servicebinding) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), sbu.waitTimeout)
	defer cancel()
	condition := sbu.isServiceBindingUsageReady()
	return resource.WaitUntilConditionSatisfied(ctx, sbu.resCli.ResCli, condition)
}

func (sbu *ServiceBindingUsage) isServiceBindingUsageReady() func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type != watch.Modified {
			return false, nil
		}
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if u.GetName() != sbu.name {
			return false, nil
		}

		servicebindingusage, err := convertFromUnstructuredToServiceBindingUsage(u)
		if err != nil {
			return false, err
		}

		return sbu.isReadyPhase(servicebindingusage), nil
	}
}

func (sbu *ServiceBindingUsage) isReadyPhase(servicebindingusage v1alpha1.ServiceBindingUsage) bool {
	if len(servicebindingusage.Status.Conditions) == 0 {
		shared.LogReadiness(false, sbu.verbose, sbu.name, sbu.log, servicebindingusage)
		return false
	}

	ready := false
	for _, condition := range servicebindingusage.Status.Conditions {
		if condition.Type == v1alpha1.ServiceBindingUsageReady && condition.Status == v1alpha1.ConditionTrue {
			ready = true
		}
	}

	shared.LogReadiness(ready, sbu.verbose, sbu.name, sbu.log, servicebindingusage)

	return ready
}

func convertFromUnstructuredToServiceBindingUsage(u *unstructured.Unstructured) (v1alpha1.ServiceBindingUsage, error) {
	servicebindingusage := v1alpha1.ServiceBindingUsage{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &servicebindingusage)
	return servicebindingusage, err
}
