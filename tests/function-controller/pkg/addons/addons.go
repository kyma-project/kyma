package addons

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"

	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type AddonConfiguration struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func New(name string, c shared.Container) *AddonConfiguration {
	return &AddonConfiguration{
		resCli:      resource.New(c.DynamicCli, v1alpha1.SchemeGroupVersion.WithResource("addonsconfigurations"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
	}
}

func (a *AddonConfiguration) Create(url string) error {
	ac := &v1alpha1.AddonsConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AddonsConfiguration",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.name,
			Namespace: a.namespace,
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{
					{
						URL: url,
					},
				},
			},
		},
	}

	_, err := a.resCli.Create(ac)
	if err != nil {
		return errors.Wrapf(err, "while creating AddonsConfigurations %s in namespace %s", a.name, a.namespace)
	}

	return err
}

func (a *AddonConfiguration) Delete() error {
	err := a.resCli.Delete(a.name, a.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while deleting AddonsConfiguration %s in namespace %s", a.name, a.namespace)
	}

	return nil
}

func (a *AddonConfiguration) Get() (*v1alpha1.AddonsConfiguration, error) {
	u, err := a.resCli.Get(a.name)
	if err != nil {
		return &v1alpha1.AddonsConfiguration{}, errors.Wrapf(err, "while getting AddonsConfiguration %s in namespace %s", a.name, a.namespace)
	}

	ac, err := convertFromUnstructuredToAddonsConfig(u)
	if err != nil {
		return &v1alpha1.AddonsConfiguration{}, err
	}

	return &ac, nil
}

func (a AddonConfiguration) LogResource() error {
	ad, err := a.Get()
	if err != nil {
		return err
	}

	out, err := helpers.PrettyMarshall(ad)
	if err != nil {
		return err
	}

	a.log.Infof("Addon Configuration Resource: %s", out)
	return nil
}

func (a *AddonConfiguration) WaitForStatusRunning() error {
	ac, err := a.Get()
	if err != nil {
		return err
	}

	// we need to ensure that status is ready first, because otherwise we would not get any events in watchtools.Until
	if a.isReadyPhase(*ac) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.waitTimeout)
	defer cancel()
	condition := a.isAddonConfigurationReady()
	_, err = watchtools.Until(ctx, ac.GetResourceVersion(), a.resCli.ResCli, condition)
	if err != nil {
		return err
	}
	return nil
}

func (a *AddonConfiguration) isAddonConfigurationReady() func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type != watch.Modified {
			return false, nil
		}
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if u.GetName() != a.name {
			return false, nil
		}

		ac, err := convertFromUnstructuredToAddonsConfig(u)
		if err != nil {
			return false, err
		}

		if ac.Status.Phase == v1alpha1.AddonsConfigurationFailed {
			return false, errors.New("Addon configuration is in failed state")
		}

		return a.isReadyPhase(ac), nil
	}
}

func (a AddonConfiguration) isReadyPhase(addonsConfig v1alpha1.AddonsConfiguration) bool {
	ready := addonsConfig.Status.CommonAddonsConfigurationStatus.Phase == v1alpha1.AddonsConfigurationReady
	shared.LogReadiness(ready, a.verbose, a.name, a.log, addonsConfig)
	return ready
}

func convertFromUnstructuredToAddonsConfig(u *unstructured.Unstructured) (v1alpha1.AddonsConfiguration, error) {
	ac := v1alpha1.AddonsConfiguration{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ac)
	return ac, err
}
