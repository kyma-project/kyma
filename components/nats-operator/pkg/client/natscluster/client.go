package natscluster

import (
	"context"
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/nats-io/nats-operator/pkg/apis/nats/v1alpha2"
)

// Client represents a NATS client.
type Client struct {
	client dynamic.Interface
}

// NewClient returns a new client instance.
func NewClient(client dynamic.Interface) *Client {
	return &Client{client}
}

// Get returns a NatsCluster given its namespace and name.
func (c *Client) Get(ctx context.Context, namespace, name string) (*v1alpha2.NatsCluster, error) {
	if obj, err := c.client.Resource(groupVersionResource()).Namespace(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
		return nil, err
	} else {
		return natsClusterFrom(obj)
	}
}

// groupVersionResource returns a new NatsCluster GroupVersionResource instance.
func groupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  v1alpha2.SchemeGroupVersion.Version,
		Group:    v1alpha2.SchemeGroupVersion.Group,
		Resource: v1alpha2.CRDResourcePlural,
	}
}

// natsClusterFrom returns a new NatsCluster instance given the unstructured object.
func natsClusterFrom(obj *unstructured.Unstructured) (*v1alpha2.NatsCluster, error) {
	data, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}
	natsCluster := new(v1alpha2.NatsCluster)
	if err = json.Unmarshal(data, natsCluster); err != nil {
		return nil, err
	}
	return natsCluster, nil
}
