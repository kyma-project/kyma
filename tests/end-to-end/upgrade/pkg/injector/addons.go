// Copied from: https://github.com/kyma-project/kyma/tree/master/tests/console-backend-service/pkg/injector
// because right now it cannot be vendored because controller-runtime and kubernetes version does not match
// we need to wait until the Helm Broker and CBS will we both at the same k8s and controller-runtime version
package injector

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/pkg/errors"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Addons provides functionality for injecting the Cluster and NS-scoped addons configuration
type Addons struct {
	client  client.Client
	url     string
	cfgName string
}

// NewAddons returns new instance of the Addons
func NewAddons(name, url string) (*Addons, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "while getting k8s client config")
	}
	cl, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, errors.Wrapf(err, "while creating k8s client")
	}

	err = v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, errors.Wrapf(err, "while registering addons configuration scheme")
	}

	return &Addons{
		client:  cl,
		url:     url,
		cfgName: name,
	}, nil
}

// InjectClusterAddonsConfiguration ensures that given addon repository is injected
func (a *Addons) InjectClusterAddonsConfiguration() error {
	ac := v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name: a.cfgName,
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{
					{
						URL: a.url,
					},
				},
			},
		},
	}

	if err := a.client.Create(context.Background(), &ac); err != nil {
		return errors.Wrap(err, "while creating addons configuration")
	}

	if err := a.ensureAddonsConfigurationIsReady(&ac); err != nil {
		return errors.Wrapf(err, "while waiting for cluster addons configuration to be ready")
	}

	return nil
}

// InjectAddonsConfiguration ensures that given addon repository is injected
func (a *Addons) InjectAddonsConfiguration(ns string) error {
	ac := v1alpha1.AddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name:      a.cfgName,
			Namespace: ns,
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{
					{
						URL: a.url,
					},
				},
			},
		},
	}

	if err := a.client.Create(context.Background(), &ac); err != nil {
		return errors.Wrap(err, "while creating addons configuration")
	}

	if err := a.ensureAddonsConfigurationIsReady(&ac); err != nil {
		return errors.Wrapf(err, "while waiting for addons configuration to be ready")
	}

	return nil
}

func (a *Addons) ensureAddonsConfigurationIsReady(obj runtime.Object) error {
	key, err := client.ObjectKeyFromObject(obj)
	if err != nil {
		return err
	}

	timeout := time.After(5 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			if err := a.client.Get(context.Background(), key, obj); err != nil {
				return errors.Wrapf(err, "fetching AddonsConfiguration %s: %v", key, err)
			}
			if a.extractStatus(obj).Phase != v1alpha1.AddonsConfigurationReady {
				return fmt.Errorf("timeout waiting to get AddonsConfiguration ready %s: %v", key, a.extractStatus(obj))
			}
		case <-tick:
			err := a.client.Get(context.Background(), key, obj)
			if err != nil {
				return errors.Wrapf(err, "Error fetching AddonsConfiguration %s: %v", key, err)
			}
			if a.extractStatus(obj).Phase == v1alpha1.AddonsConfigurationReady {
				return nil
			}
		}
	}
}

// CleanupAddonsConfiguration ensures that given addon repository is removed
func (a *Addons) CleanupAddonsConfiguration(ns string) error {
	ac := v1alpha1.AddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name:      a.cfgName,
			Namespace: ns,
		},
	}
	return a.delete(&ac)
}

// CleanupClusterAddonsConfiguration ensures that given addon repository is removed
func (a *Addons) CleanupClusterAddonsConfiguration() error {
	ac := v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name: a.cfgName,
		},
	}
	return a.delete(&ac)
}

func (a *Addons) delete(obj runtime.Object) error {
	key, err := client.ObjectKeyFromObject(obj)
	if err != nil {
		return err
	}

	err = a.client.Delete(context.Background(), obj)
	if err != nil && !apierror.IsNotFound(err) {
		return errors.Wrapf(err, "while deleting the addons configuration %s: %v", a.cfgName, err)
	}

	err = wait.PollImmediate(time.Second, time.Minute, func() (done bool, err error) {
		err = a.client.Get(context.Background(), key, obj)

		if apierror.IsNotFound(err) {
			return true, nil
		}

		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("timeout occured when waiting for %s to be deleted", key.String())
	}

	return err
}

func (*Addons) extractStatus(object runtime.Object) v1alpha1.CommonAddonsConfigurationStatus {
	switch t := object.(type) {
	case *v1alpha1.ClusterAddonsConfiguration:
		return t.Status.CommonAddonsConfigurationStatus
	case *v1alpha1.AddonsConfiguration:
		return t.Status.CommonAddonsConfigurationStatus
	}

	return v1alpha1.CommonAddonsConfigurationStatus{}
}
