package assetstore

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	DocsTopicModeSingle = "single"
	ApiSpecKey          = "api"
	EventSpecKey        = "events"
	DocsSpecKey         = "docs"
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

	sources := make(map[string]v1alpha1.Source)
	for key, url := range docsTopicEntry.Urls {
		sources[key] = v1alpha1.Source{
			URL:  url,
			Mode: DocsTopicModeSingle,
		}
	}

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

	unstructured, err := r.resourceInterface.Get(id, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return docstopic.Entry{}, apperrors.NotFound("Docs Topic with id:%s not found", id)
		}

		if err != nil {
			return docstopic.Entry{}, nil
		}
	}

	var docsTopic v1alpha1.DocsTopic
	runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, &docsTopic)
	if err != nil {
		return docstopic.Entry{}, apperrors.Internal("Failed to convert from unstructured object.")
	}

	return fromK8sType(docsTopic), nil
}

func fromK8sType(k8sDocsTopic v1alpha1.DocsTopic) docstopic.Entry {
	urls := make(map[string]string)
	for key, source := range k8sDocsTopic.Spec.Sources {
		urls[key] = source.URL
	}

	return docstopic.Entry{
		Id:          k8sDocsTopic.Name,
		Description: k8sDocsTopic.Spec.Description,
		DisplayName: k8sDocsTopic.Spec.DisplayName,
		Urls:        urls,
		Labels:      k8sDocsTopic.Labels,
	}
}

func (r repository) Delete(id string) apperrors.AppError {
	return nil
}

func (r repository) Update(documentationTopicdocstopic docstopic.Entry) apperrors.AppError {
	return nil
}
