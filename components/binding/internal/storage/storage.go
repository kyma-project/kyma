package storage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceData struct {
	Schema      schema.GroupVersionResource
	LabelsPath  string
	LabelFields []string
}

type KindStorage struct {
	registered map[v1alpha1.Kind]*ResourceData
	mu         sync.RWMutex
}

func NewKindStorage() *KindStorage {
	return &KindStorage{
		registered: make(map[v1alpha1.Kind]*ResourceData),
	}
}

func newResourceData(gvr schema.GroupVersionResource, labelsPath string) *ResourceData {
	return &ResourceData{
		Schema:      gvr,
		LabelsPath:  labelsPath,
		LabelFields: strings.Split(labelsPath, "."),
	}
}

// Register adds KindManager for given Kind to KindStorage
func (s *KindStorage) Register(tk v1alpha1.TargetKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registered[tk.Spec.Resource.Kind] = newResourceData(schema.GroupVersionResource{
		Group:    tk.Spec.Resource.Group,
		Version:  tk.Spec.Resource.Version,
		Resource: strings.ToLower(fmt.Sprintf("%s%s", tk.Spec.Resource.Kind, "s")),
	}, tk.Spec.LabelsPath)

	return nil
}

// Unregister removes given Kind from KindStorage
func (s *KindStorage) Unregister(tk v1alpha1.TargetKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.registered, tk.Spec.Resource.Kind)
	return nil
}

// Get returns KindManager for given Kind
func (s *KindStorage) Get(kind v1alpha1.Kind) (*ResourceData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	concreteResourceData, exists := s.registered[kind]
	if !exists {
		return &ResourceData{}, fmt.Errorf("TargetKind %s was not found", kind)
	}
	return concreteResourceData, nil
}

func (s *KindStorage) Exist(tk v1alpha1.TargetKind) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.registered[tk.Spec.Resource.Kind]
	if !exists {
		return false
	}
	return exists
}

func (s *KindStorage) Equal(tk v1alpha1.TargetKind, registeredTk *ResourceData) bool {
	if tk.Spec.Resource.Group != registeredTk.Schema.Group || fmt.Sprintf("%s%s", tk.Spec.Resource.Kind, "s") != registeredTk.Schema.Resource || tk.Spec.Resource.Version != registeredTk.Schema.Version || tk.Spec.LabelsPath != registeredTk.LabelsPath {
		return false
	}
	return true
}
