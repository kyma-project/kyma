package serviceinstance

import (
	"context"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceInstance struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func (si *ServiceInstance) GetName() string {
	return si.name
}

func New(name string, c shared.Container) *ServiceInstance {
	return &ServiceInstance{
		resCli:      resource.New(c.DynamicCli, v1beta1.SchemeGroupVersion.WithResource("serviceinstances"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
	}
}

func (si *ServiceInstance) Create(serviceClassExternalName, servicePlanExternalName string) error {
	ac := &v1beta1.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceInstance",
			APIVersion: v1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      si.name,
			Namespace: si.namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassExternalName: serviceClassExternalName,
				ServicePlanExternalName:  servicePlanExternalName,
			},
			Parameters: &runtime.RawExtension{
				Raw: []byte(`{"imagePullPolicy": "Always"}`),
			},
		},
	}

	_, err := si.resCli.Create(ac)
	if err != nil {
		return errors.Wrapf(err, "while creating ServiceInstance %s in namespace %s", si.name, si.namespace)
	}

	return err
}

func (si *ServiceInstance) Delete() error {
	err := si.resCli.Delete(si.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ServiceInstance %s in namespace %s", si.name, si.namespace)
	}

	return nil
}

func (si *ServiceInstance) Get() (*v1beta1.ServiceInstance, error) {
	u, err := si.resCli.Get(si.name)
	if err != nil {
		return &v1beta1.ServiceInstance{}, errors.Wrapf(err, "while getting ServiceInstance %s in namespace %s", si.name, si.namespace)
	}

	serviceinstance, err := convertFromUnstructuredToServiceInstance(u)
	if err != nil {
		return &v1beta1.ServiceInstance{}, err
	}

	return &serviceinstance, nil
}

func (si *ServiceInstance) LogResource() error {
	serviceInstance, err := si.Get()
	if err != nil {
		return err
	}

	out, err := helpers.PrettyMarshall(serviceInstance)
	if err != nil {
		return err
	}

	si.log.Infof("%s", out)
	return nil
}

func (si *ServiceInstance) WaitForStatusRunning() error {
	serviceinstance, err := si.Get()
	if err != nil {
		return err
	}

	// we need to ensure that status is ready first, because otherwise we would not Get any events in watchtools.Until
	if si.isReadyPhase(*serviceinstance) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), si.waitTimeout)
	defer cancel()
	condition := si.isServiceInstanceReady()
	return resource.WaitUntilConditionSatisfied(ctx, si.resCli.ResCli, condition)
}

func (si *ServiceInstance) isServiceInstanceReady() func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type != watch.Modified {
			return false, nil
		}
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if u.GetName() != si.name {
			return false, nil
		}

		serviceinstance, err := convertFromUnstructuredToServiceInstance(u)
		if err != nil {
			return false, err
		}

		return si.isReadyPhase(serviceinstance), nil
	}
}

func (si ServiceInstance) isReadyPhase(serviceinstance v1beta1.ServiceInstance) bool {
	if len(serviceinstance.Status.Conditions) == 0 {
		shared.LogReadiness(false, si.verbose, si.name, si.log, serviceinstance)
		return false
	}

	ready := false
	for _, condition := range serviceinstance.Status.Conditions {
		if condition.Type == v1beta1.ServiceInstanceConditionReady && condition.Status == v1beta1.ConditionTrue {
			ready = true
		}
	}
	shared.LogReadiness(ready, si.verbose, si.name, si.log, serviceinstance)

	return ready
}

func convertFromUnstructuredToServiceInstance(u *unstructured.Unstructured) (v1beta1.ServiceInstance, error) {
	si := v1beta1.ServiceInstance{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &si)
	return si, err
}
