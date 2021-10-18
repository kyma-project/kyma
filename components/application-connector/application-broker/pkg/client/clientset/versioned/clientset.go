// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-connector/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	ApplicationconnectorV1alpha1() applicationconnectorv1alpha1.ApplicationconnectorV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	applicationconnectorV1alpha1 *applicationconnectorv1alpha1.ApplicationconnectorV1alpha1Client
}

// ApplicationconnectorV1alpha1 retrieves the ApplicationconnectorV1alpha1Client
func (c *Clientset) ApplicationconnectorV1alpha1() applicationconnectorv1alpha1.ApplicationconnectorV1alpha1Interface {
	return c.applicationconnectorV1alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.applicationconnectorV1alpha1, err = applicationconnectorv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.applicationconnectorV1alpha1 = applicationconnectorv1alpha1.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.applicationconnectorV1alpha1 = applicationconnectorv1alpha1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
