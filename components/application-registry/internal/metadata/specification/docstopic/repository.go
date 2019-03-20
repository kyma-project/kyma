package docstopic

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ResourceInterface interface {
	// Get gets the resource with the specified name.
	Get(name string, opts metav1.GetOptions) (*unstructured.Unstructured, error)
	// Delete deletes the resource with the specified name.
	Delete(name string, opts *metav1.DeleteOptions) error
	// Create creates the provided resource.
	Create(obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
	// Update updates the provided resource.
	Update(obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

type SpecEntry struct {
	Url string
	Key string
}

type Entry struct {
	Id            string
	DisplayName   string
	Description   string
	ApiSpec       *SpecEntry
	EventsSpec    *SpecEntry
	Documentation *SpecEntry
}

type Repository interface {
	Create(documentationTopic Entry) apperrors.AppError
	Get(id string) (Entry, apperrors.AppError)
	Delete(id string) apperrors.AppError
	Update(documentationTopic Entry) apperrors.AppError
}

type repository struct {
}

func NewRepository(resourceInterface ResourceInterface) Repository {
	return repository{}
}

func (r repository) Create(documentationTopic Entry) apperrors.AppError {

	return nil
}

func (r repository) Get(id string) (Entry, apperrors.AppError) {
	return Entry{}, nil
}

func (r repository) Delete(id string) apperrors.AppError {
	return nil
}

func (r repository) Update(documentationTopic Entry) apperrors.AppError {
	return nil
}
