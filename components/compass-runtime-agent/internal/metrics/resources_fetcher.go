package metrics

import (
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ResourcesFetcher interface {
	FetchNodesResources() ([]NodeResources, error)
}

type resourcesFetcher struct {
	nodeClientSet core.NodeInterface
}

func newResourcesFetcher(clientset kubernetes.Interface) ResourcesFetcher {
	return &resourcesFetcher{
		nodeClientSet: clientset.CoreV1().Nodes(),
	}
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
