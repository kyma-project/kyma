package webhook

import (
	"context"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	defaultInterval = 500 * time.Millisecond
	defaultTimeout  = 3 * time.Second
)

func NewClient(c cli.Client) Client {
	return &client{
		client: c,
	}
}

type Client interface {
	FindSecret(ctx context.Context, binding *v1alpha1.Binding) (*corev1.Secret, error)
	FindConfigMap(ctx context.Context, binding *v1alpha1.Binding) (*corev1.ConfigMap, error)
	FindBindings(ctx context.Context, bindingsName []string, namespace string) ([]*v1alpha1.Binding, error)
}

type client struct {
	client cli.Client
}

// findSecret finds Secret based on Binding Source field
func (c *client) FindSecret(ctx context.Context, binding *v1alpha1.Binding) (*corev1.Secret, error) {
	secret := &corev1.Secret{}

	var lastError error
	err := wait.PollImmediate(defaultInterval, defaultTimeout, func() (bool, error) {
		err := c.client.Get(ctx, cli.ObjectKey{Name: binding.Spec.Source.Name, Namespace: binding.Namespace}, secret)
		if err != nil {
			lastError = err
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return secret, errors.Wrapf(lastError, "while getting Secret %s/%s", binding.Namespace, binding.Spec.Source.Name)
	}

	return secret, nil
}

// findConfigMap finds ConfigMap based on Binding Source field
func (c *client) FindConfigMap(ctx context.Context, binding *v1alpha1.Binding) (*corev1.ConfigMap, error) {
	configmap := &corev1.ConfigMap{}

	var lastError error
	err := wait.PollImmediate(defaultInterval, defaultTimeout, func() (bool, error) {
		err := c.client.Get(ctx, cli.ObjectKey{Name: binding.Spec.Source.Name, Namespace: binding.Namespace}, configmap)
		if err != nil {
			lastError = err
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return configmap, errors.Wrapf(lastError, "while getting ConfigMap %s/%s", binding.Namespace, binding.Spec.Source.Name)
	}

	return configmap, nil
}

// findBindings fetches all Bindings based on Bindings name and request namespace
func (c *client) FindBindings(ctx context.Context, bindingsName []string, namespace string) ([]*v1alpha1.Binding, error) {
	bindings := make([]*v1alpha1.Binding, 0)

	for _, bindingName := range bindingsName {
		var binding = &v1alpha1.Binding{}
		var lastError error
		err := wait.PollImmediate(defaultInterval, defaultTimeout, func() (bool, error) {
			err := c.client.Get(ctx, cli.ObjectKey{Name: bindingName, Namespace: namespace}, binding)
			if err != nil {
				lastError = err
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			return bindings, errors.Wrapf(lastError, "while getting Binding %s/%s", bindingName, namespace)
		}
		bindings = append(bindings, binding)
	}

	return bindings, nil
}
