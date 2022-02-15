package uaa

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/uaa-activator/internal/repeat"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Creator provides functionality for creating UAA Instance and Binding
type Creator struct {
	cli    client.Client
	config Config
}

// NewCreator returns new instance of Creator
func NewCreator(cli client.Client, config Config) *Creator {
	return &Creator{
		cli:    cli,
		config: config,
	}
}

// EnsureUAAInstance ensures that ServiceInstance is created and up to date.
// Additionally, wait until ServiceInstance is in a ready state.
func (p *Creator) EnsureUAAInstance(ctx context.Context) error {
	instance := p.uaaServiceInstance()

	// remove not ready ServiceInstance before create new one
	if err := p.instanceIsReady(ctx); err != nil {
		err := p.removeUAAInstance(ctx, instance)
		if err != nil {
			return errors.Wrap(err, "while removing ServiceInstance in not ready state")
		}
		// there is a problem with communication with the UAA. If an instance after deletion is created
		// in too short a period of time, the UAA provider will return information that the instance with
		// the given 'xsappname' is in the deleting process and the new instance will not be triggered to create it
		// so there is a delay before a request for a new instance is sent
		time.Sleep(p.config.NewInstanceCreateDelay)
	}

	err := p.cli.Create(ctx, &instance)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			old := v1beta1.ServiceInstance{}
			if err := p.cli.Get(ctx, p.config.ServiceInstance, &old); err != nil {
				return errors.Wrap(err, "while fetching service instance")
			}

			// Updating only PlanReference and Parameters, other fields are immutable
			// and need to remain the same, otherwise you get such an error:
			//   admission webhook "validating.serviceinstances.servicecatalog.k8s.io" denied the request: spec.externalID: Invalid value: "": field is immutable
			toUpdate := old.DeepCopy()
			toUpdate.Spec.PlanReference = instance.Spec.PlanReference
			toUpdate.Spec.ParametersFrom = instance.Spec.ParametersFrom
			return p.cli.Update(ctx, toUpdate)
		})
		if err != nil {
			return errors.Wrap(err, "while updating UAA ServiceInstance")
		}
	default:
		return err
	}

	err = repeat.UntilSuccess(ctx, p.instanceIsReady(ctx))
	if err != nil {
		return errors.Wrapf(err, "while waiting for %s", p.config.ServiceInstance.String())
	}

	return nil
}

// EnsureUAABinding ensures that ServiceBinding is created and up to date.
// Additionally, wait until ServiceBinding is in a ready state.
func (p *Creator) EnsureUAABinding(ctx context.Context) error {
	binding := p.uaaServiceBinding()

	err := p.cli.Create(ctx, &binding)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
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

	return nil
}

func (p *Creator) uaaServiceInstance() v1beta1.ServiceInstance {
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
			ParametersFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &p.config.ServiceInstanceParamsSecret,
				},
			},
		},
	}
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

func (p *Creator) removeUAAInstance(ctx context.Context, instance v1beta1.ServiceInstance) error {
	err := p.cli.Delete(ctx, &instance)
	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		return nil
	case err != nil:
		return errors.Wrap(err, "while deleting UAA ServiceInstance")
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
