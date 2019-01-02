// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/kyma-project/kyma/components/connector-service/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeApplicationconnectorV1alpha1 struct {
	*testing.Fake
}

func (c *FakeApplicationconnectorV1alpha1) KymaGroups() v1alpha1.KymaGroupInterface {
	return &FakeKymaGroups{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeApplicationconnectorV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
