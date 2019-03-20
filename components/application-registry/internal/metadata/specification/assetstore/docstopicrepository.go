package assetstore

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	DocsTopicModeSingle = "single"
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
	Create(documentationTopic docstopic.Entry) apperrors.AppError
	Get(id string) (docstopic.Entry, apperrors.AppError)
	Delete(id string) apperrors.AppError
	Update(documentationTopic docstopic.Entry) apperrors.AppError
}

type repository struct {
	resourceInterface ResourceInterface
	namespace         string
}

func NewDocsTopicRepository(resourceInterface ResourceInterface, namespace string) DocsTopicRepository {
	return repository{
		resourceInterface: resourceInterface,
		namespace:         namespace,
	}
}

func (r repository) Create(documentationTopic docstopic.Entry) apperrors.AppError {

	docsTopic := toK8sType(documentationTopic, r.namespace)
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&docsTopic)
	if err != nil {
		return apperrors.Internal("Failed to convert Docs Topic object.")
	}

	_, err = r.resourceInterface.Create(&unstructured.Unstructured{
		Object: obj,
	})

	if err != nil {
		return apperrors.Internal("Failed to create Documentation Topic")
	}

	return nil
}

func toK8sType(docsTopicEntry docstopic.Entry, namespace string) v1alpha1.DocsTopic {

	addSource := func(entry *docstopic.SpecEntry, sources map[string]v1alpha1.Source) {
		if entry != nil {
			source := v1alpha1.Source{
				URL:  entry.Url,
				Mode: DocsTopicModeSingle,
			}
			sources[entry.Key] = source
		}
	}

	sources := make(map[string]v1alpha1.Source)
	addSource(docsTopicEntry.ApiSpec, sources)
	addSource(docsTopicEntry.EventsSpec, sources)
	addSource(docsTopicEntry.Documentation, sources)

	return v1alpha1.DocsTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:      docsTopicEntry.Id,
			Namespace: namespace,
			Labels:    docsTopicEntry.Labels,
		},
		Spec: v1alpha1.DocsTopicSpec{
			CommonDocsTopicSpec: v1alpha1.CommonDocsTopicSpec{
				DisplayName: "Some display name",
				Description: "Some description",
				Sources:     sources,
			},
		}}
}

func (r repository) Get(id string) (docstopic.Entry, apperrors.AppError) {
	return docstopic.Entry{}, nil
}

func (r repository) Delete(id string) apperrors.AppError {
	return nil
}

func (r repository) Update(documentationTopicdocstopic docstopic.Entry) apperrors.AppError {
	return nil
}
