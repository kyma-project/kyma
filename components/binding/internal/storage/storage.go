package storage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type KindStorage interface {
	Register(tk v1alpha1.TargetKind) error
	Unregister(tk v1alpha1.TargetKind) error
	Get(kind v1alpha1.Kind) (*ResourceData, error)
	Exist(tk v1alpha1.TargetKind) bool
	Equal(tk v1alpha1.TargetKind, registeredTk *ResourceData) bool
}

type storage struct {
	registered map[v1alpha1.Kind]*ResourceData
	mu         sync.RWMutex
}

func NewKindStorage() KindStorage {
	return &storage{
		registered: make(map[v1alpha1.Kind]*ResourceData),
	}
}

// Register adds KindManager for given Kind to storage
func (s *storage) Register(tk v1alpha1.TargetKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registered[tk.Spec.Resource.Kind] = newResourceData(schema.GroupVersionResource{
		Group:    tk.Spec.Resource.Group,
		Version:  tk.Spec.Resource.Version,
		Resource: strings.ToLower(fmt.Sprintf("%s%s", tk.Spec.Resource.Kind, "s")),
	}, tk.Spec.LabelsPath)

	return nil
}

// Unregister removes given Kind from storage
func (s *storage) Unregister(tk v1alpha1.TargetKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.registered, tk.Spec.Resource.Kind)
	return nil
}

// Get returns KindManager for given Kind
func (s *storage) Get(kind v1alpha1.Kind) (*ResourceData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	concreteResourceData, exists := s.registered[kind]
	if !exists {
		return &ResourceData{}, fmt.Errorf("TargetKind %s was not found", kind)
	}
	return concreteResourceData, nil
}

func (s *storage) Exist(tk v1alpha1.TargetKind) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.registered[tk.Spec.Resource.Kind]
	if !exists {
		return false
	}
	return exists
}

func (s *storage) Equal(tk v1alpha1.TargetKind, registeredTk *ResourceData) bool {
	if tk.Spec.Resource.Group != registeredTk.Schema.Group || fmt.Sprintf("%s%s", tk.Spec.Resource.Kind, "s") != registeredTk.Schema.Resource || tk.Spec.Resource.Version != registeredTk.Schema.Version || tk.Spec.LabelsPath != registeredTk.LabelsPath {
		return false
	}
	return true
}

func newResourceData(gvr schema.GroupVersionResource, labelsPath string) *ResourceData {
	return &ResourceData{
		Schema:      gvr,
		LabelsPath:  labelsPath,
		LabelFields: strings.Split(labelsPath, "."),
	}
}
