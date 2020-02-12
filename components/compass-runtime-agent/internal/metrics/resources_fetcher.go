package metrics

import (
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type ResourcesFetcher interface {
	FetchNodesResources() ([]NodeResources, error)
}

type resourcesFetcher struct {
	nodeClientSet core.NodeInterface
}

func newResourcesFetcher(config *rest.Config) (ResourcesFetcher, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create clientset for config")
	}

	return &resourcesFetcher{
		nodeClientSet: clientset.CoreV1().Nodes(),
	}, nil
}

func (r *resourcesFetcher) FetchNodesResources() ([]NodeResources, error) {
	nodes, err := r.nodeClientSet.List(meta.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list nodes")
	}

	var clusterResources []NodeResources

	for _, node := range nodes.Items {
		clusterResources = append(clusterResources, NodeResources{
			Name:         node.Name,
			InstanceType: node.Labels["beta.kubernetes.io/instance-type"],
			Capacity: ResourceGroup{
				CPU:              node.Status.Capacity.Cpu().String(),
				EphemeralStorage: node.Status.Capacity.StorageEphemeral().String(),
				Memory:           node.Status.Capacity.Memory().String(),
				Pods:             node.Status.Capacity.Pods().String(),
			},
		})
	}

	return clusterResources, nil
}
