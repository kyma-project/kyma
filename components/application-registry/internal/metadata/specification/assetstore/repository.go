package assetstore

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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

type DocsTopicRepository interface {
	Create(docstopic *v1alpha1.DocsTopic) (*v1alpha1.DocsTopic, error)
	Get(name string, options v1.GetOptions) (*v1alpha1.DocsTopic, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Update(docstopic *v1alpha1.DocsTopic) (*v1alpha1.DocsTopic, error)
}

type ApiType int

const (
	Swagger   ApiType = 0
	ODataXML  ApiType = 1
	ODataJSON ApiType = 2
)

type ApiSpec struct {
	Url  string
	Type ApiType
}

type EventsSpec struct {
	Url string
}

type Documentation struct {
	Url string
}

type DocumentationTopic struct {
	id            string
	ApiSpec       *ApiSpec
	EventsSpec    *EventsSpec
	Documentation *Documentation
}

type Repository interface {
	Create(documentationTopic DocumentationTopic) apperrors.AppError
	Get(id string) (DocumentationTopic, apperrors.AppError)
	Delete(id string) apperrors.AppError
	Update(documentationTopic DocumentationTopic) apperrors.AppError
}

type repository struct {
}

func NewRepository(resourceInterface ResourceInterface) Repository {
	return repository{}
}

func (r repository) Create(documentationTopic DocumentationTopic) apperrors.AppError {
	return nil
}

func (r repository) Get(id string) (DocumentationTopic, apperrors.AppError) {
	return DocumentationTopic{}, nil
}

func (r repository) Delete(id string) apperrors.AppError {
	return nil
}

func (r repository) Update(documentationTopic DocumentationTopic) apperrors.AppError {
	return nil
}
