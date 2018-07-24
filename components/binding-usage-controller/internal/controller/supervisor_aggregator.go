package controller

import "fmt"

//go:generate mockery -name=KubernetesResourceSupervisor -output=automock -outpkg=automock -case=underscore

// KubernetesResourceSupervisor validates if given Kubernetes resource can be modified by ServiceBindingUsage. If yes
// it can ensure that labels are present or deleted on previous validated resource.
type KubernetesResourceSupervisor interface {
	HasSynced() bool
	EnsureLabelsCreated(resourceNs, resourceName, usageName string, labels map[string]string) error
	EnsureLabelsDeleted(resourceNs, resourceName, usageName string) error
	GetInjectedLabels(resourceNs, resourceName, usageName string) (map[string]string, error)
}

// Kind represents Kubernetes Kind name
type Kind string

const (
	// KindDeployment represents Deployment resource
	KindDeployment Kind = "Deployment"
	// KindKubelessFunction represents Kubeless Function resource
	KindKubelessFunction Kind = "Function"
)

// ResourceSupervisorAggregator aggregates all defined resources supervisors
type ResourceSupervisorAggregator struct {
	registered map[Kind]KubernetesResourceSupervisor
}

// NewResourceSupervisorAggregator returns new instance of ResourceSupervisorAggregator
func NewResourceSupervisorAggregator() *ResourceSupervisorAggregator {
	return &ResourceSupervisorAggregator{
		registered: make(map[Kind]KubernetesResourceSupervisor),
	}
}

// Register adds new resource supervisor
func (f *ResourceSupervisorAggregator) Register(k Kind, supervisor KubernetesResourceSupervisor) error {
	if _, exists := f.registered[k]; exists {
		return fmt.Errorf("supervisor for kind %q is already registered", k)
	}

	f.registered[k] = supervisor
	return nil
}

// HasSynced returns true if all registered supervisors are synced
func (f *ResourceSupervisorAggregator) HasSynced() bool {
	for _, supervisor := range f.registered {
		if !supervisor.HasSynced() {
			return false
		}
	}

	return true
}

// Get returns supervisor for given kind
func (f *ResourceSupervisorAggregator) Get(k Kind) (KubernetesResourceSupervisor, error) {
	concreteSupervisor, exists := f.registered[k]
	if !exists {
		return nil, NewNotFoundError("supervisor for kind %s was not found", k)
	}

	return concreteSupervisor, nil
}
