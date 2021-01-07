package uaa

import (
	"context"

	"github.com/kyma-project/kyma/components/uaa-activator/internal/repeat"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Creator provides functionality for creating UAA Instance and Binding
type Creator struct {
	cli        client.Client
	config     Config
	parameters *ParametersBuilder
	log        *zap.SugaredLogger
}

// NewCreator returns new instance of Creator
func NewCreator(cli client.Client, config Config, domain string, log *zap.SugaredLogger) *Creator {
	return &Creator{
		cli:        cli,
		config:     config,
		parameters: NewParametersBuilder(config, domain),
		log:        log,
	}
}

// EnsureUAAInstance ensures that ServiceInstance is created and up to date.
// Additionally, wait until ServiceInstance is in a ready state.
func (p *Creator) EnsureUAAInstance(ctx context.Context) error {
	if p.config.IsUpgrade {
		return p.processUpgradeInstance(ctx)
	} else {
		return p.processInstallInstance(ctx)
	}
}

func (p *Creator) processInstallInstance(ctx context.Context) error {
	p.log.Info("start ServiceInstance install process")

	// check if ServiceInstance exist and is ready,
	// if exist and is not ready try remove
	exist, err := p.instanceExist(ctx)
	if err != nil {
		return errors.Wrap(err, "while checking if ServiceInstance exist")
	}
	if exist {
		f := p.instanceIsReady(ctx)
		if err := f(); err == nil {
			p.log.Info("ServiceInstance exist and is in ready state, install process successfully completed")
			return nil
		}
		p.log.Info("ServiceInstance exist and is in not ready state, try remove instance")
		err = p.removeUAAInstance(ctx)
		if err != nil {
			return errors.Wrap(err, "while removing not ready ServiceInstance")
		}
	}

	// create new instance with random xsappname
	p.log.Info("create new ServiceInstance")
	instance, err := p.uaaServiceInstance(nil)
	if err != nil {
		return errors.Wrap(err, "while creating ServiceInstance schema")
	}
	err = p.cli.Create(ctx, &instance)
	if err != nil {
		return errors.Wrap(err, "while creating ServiceInstance")
	}

	p.log.Info("wait until ServiceInstance will be in ready state")
	err = repeat.UntilSuccess(ctx, p.instanceIsReady(ctx))
	if err != nil {
		return errors.Wrapf(err, "while waiting for %s", p.config.ServiceInstance.String())
	}

	p.log.Info("ServiceInstance install process successfully completed")
	return nil
}

func (p *Creator) processUpgradeInstance(ctx context.Context) error {
	p.log.Info("start ServiceInstance upgrade process")

	// check if ServiceInstance exist and is ready, return error if not
	// ServiceInstance during upgrade process has to exist and be ready
	exist, err := p.instanceExist(ctx)
	if err != nil {
		return errors.Wrap(err, "while checking if ServiceInstance exist")
	}
	if !exist {
		return errors.New("ServiceInstance not exist in upgrade process")
	}

	f := p.instanceIsReady(ctx)
	if err := f(); err != nil {
		return errors.New("ServiceInstance is not in ready state in upgrade process")
	}

	// create instance with parameters with the same xsappname
	// and update existing ServiceInstance
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		old := v1beta1.ServiceInstance{}
		if err := p.cli.Get(ctx, p.config.ServiceInstance, &old); err != nil {
			return errors.Wrap(err, "while fetching service instance")
		}
		instance := old.DeepCopy()
		toUpdate, err := p.uaaServiceInstance(instance)
		if err != nil {
			return errors.Wrap(err, "while updating ServiceInstance")
		}

		p.log.Info("update ServiceInstance in upgrade process")
		instance.Spec.PlanReference = toUpdate.Spec.PlanReference
		instance.Spec.ParametersFrom = toUpdate.Spec.ParametersFrom
		instance.Spec.Parameters = toUpdate.Spec.Parameters

		return p.cli.Update(ctx, instance)
	})
	if err != nil {
		return errors.Wrap(err, "while updating UAA ServiceInstance")
	}

	p.log.Info("wait until ServiceInstance will be in ready state")
	err = repeat.UntilSuccess(ctx, p.instanceIsReady(ctx))
	if err != nil {
		return errors.Wrapf(err, "while waiting for %s", p.config.ServiceInstance.String())
	}

	p.log.Info("ServiceInstance upgrade process successfully completed")
	return nil
}

// EnsureUAABinding ensures that ServiceBinding is created and up to date.
// Additionally, wait until ServiceBinding is in a ready state.
func (p *Creator) EnsureUAABinding(ctx context.Context) error {
	p.log.Info("start ServiceBinding process")
	binding := p.uaaServiceBinding()

	err := p.cli.Create(ctx, &binding)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		p.log.Info("ServiceBinding already exist, trying update")
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			old := v1beta1.ServiceBinding{}
			if err := p.cli.Get(ctx, p.config.ServiceBinding, &old); err != nil {
				return err
			}

			toUpdate := old.DeepCopy()
			toUpdate.Spec = binding.Spec
			return p.cli.Update(ctx, toUpdate)
		})
		if err != nil {
			return errors.Wrap(err, "while updating UAA ServiceBinding")
		}
	default:
		return err
	}

	err = repeat.UntilSuccess(ctx, p.bindingIsReady(ctx))
	if err != nil {
		return errors.Wrapf(err, "while waiting for %s", p.config.ServiceBinding.String())
	}

	err = repeat.UntilSuccess(ctx, p.secretExist(ctx))
	if err != nil {
		return errors.Wrapf(err, "while waiting for Secret %s", p.config.ServiceBinding.String())
	}

	p.log.Info("ServiceBinding process successfully completed")
	return nil
}

func (p *Creator) uaaServiceInstance(instance *v1beta1.ServiceInstance) (v1beta1.ServiceInstance, error) {
	parameters, err := p.parameters.Generate(instance)
	if err != nil {
		return v1beta1.ServiceInstance{}, errors.Wrap(err, "while creating instance parameters")
	}

	return v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.config.ServiceInstance.Name,
			Namespace: p.config.ServiceInstance.Namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassName: p.config.ClusterServiceClassName,
				ClusterServicePlanName:  p.config.ClusterServicePlanName,
			},
			Parameters: &runtime.RawExtension{Raw: parameters},
		},
	}, nil
}

func (p *Creator) uaaServiceBinding() v1beta1.ServiceBinding {
	return v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.config.ServiceBinding.Name,
			Namespace: p.config.ServiceBinding.Namespace,
		},
		Spec: v1beta1.ServiceBindingSpec{
			InstanceRef: v1beta1.LocalObjectReference{
				Name: p.config.ServiceInstance.Name,
			},
		},
	}
}

func (p *Creator) removeUAAInstance(ctx context.Context) error {
	err := repeat.UntilSuccess(ctx, func() (err error) {
		instance := &v1beta1.ServiceInstance{}

		err = p.cli.Get(ctx, p.config.ServiceInstance, instance)
		switch {
		case err == nil:
		case apiErrors.IsNotFound(err):
			return nil
		case err != nil:
			return errors.Wrap(err, "while getting UAA ServiceInstance")
		}

		err = p.cli.Delete(ctx, instance)
		if err != nil && !apiErrors.IsNotFound(err) {
			return errors.Wrap(err, "while deleting UAA ServiceInstance")
		}
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "while removing %s", p.config.ServiceInstance.String())
	}

	err = repeat.UntilSuccess(ctx, p.instanceWasRemoved(ctx))
	if err != nil {
		return errors.Wrapf(err, "while waiting for removing %s", p.config.ServiceInstance.String())
	}

	return nil
}

func (p *Creator) instanceIsReady(ctx context.Context) func() error {
	return func() error {
		instance := v1beta1.ServiceInstance{}
		if err := p.cli.Get(ctx, p.config.ServiceInstance, &instance); err != nil {
			return err
		}

		for _, cond := range instance.Status.Conditions {
			if cond.Type == v1beta1.ServiceInstanceConditionReady &&
				cond.Status == v1beta1.ConditionTrue {
				return nil
			}
			p.log.Infof("[SI debug] ServiceInstance condition message: %s", cond.Message)
		}

		return errors.Errorf("ServiceInstance is not ready, status: %v", instance.Status)
	}
}

func (p *Creator) instanceWasRemoved(ctx context.Context) func() error {
	return func() error {
		instance := v1beta1.ServiceInstance{}
		err := p.cli.Get(ctx, p.config.ServiceInstance, &instance)

		if apiErrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "while checking if ServiceInstance was removed")
		}

		return errors.Errorf("ServiceInstance exist, status: %v", instance.Status)
	}
}

func (p *Creator) bindingIsReady(ctx context.Context) func() error {
	return func() error {
		binding := v1beta1.ServiceBinding{}
		if err := p.cli.Get(ctx, p.config.ServiceBinding, &binding); err != nil {
			return err
		}

		for _, cond := range binding.Status.Conditions {
			if cond.Type == v1beta1.ServiceBindingConditionReady &&
				cond.Status == v1beta1.ConditionTrue {
				return nil
			}
			p.log.Infof("[SB debug] ServiceBinding condition message: %s", cond.Message)
		}

		return errors.Errorf("ServiceBinding is not ready, status: %v", binding.Status)
	}
}

func (p *Creator) secretExist(ctx context.Context) func() error {
	return func() error {
		secret := v1.Secret{}
		err := p.cli.Get(ctx, p.config.ServiceBinding, &secret)
		if err != nil {
			return errors.Wrapf(err, "while fetching Secret %s", p.config.ServiceBinding.Name)
		}

		return nil
	}
}

func (p *Creator) instanceExist(ctx context.Context) (bool, error) {
	instance := v1beta1.ServiceInstance{}
	err := p.cli.Get(ctx, p.config.ServiceInstance, &instance)
	switch {
	case err == nil:
		return true, nil
	case apiErrors.IsNotFound(err):
		return false, nil
	default:
		return false, err
	}
}
