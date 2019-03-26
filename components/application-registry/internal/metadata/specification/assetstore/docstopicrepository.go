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
)

type ResourceInterface interface {
	// Get gets the resource with the specified name.
	Get(name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
	// Delete deletes the resource with the specified name.
	Delete(name string, opts *metav1.DeleteOptions, subresources ...string) error
	// Create creates the provided resource.
	Create(obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error)
	// Update updates the provided resource.
	Update(obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error)
}

type DocsTopicRepository interface {
	Get(id string) (docstopic.Entry, apperrors.AppError)
	Upsert(documentationTopic docstopic.Entry) apperrors.AppError
	Delete(id string) apperrors.AppError
}

type repository struct {
	resourceInterface ResourceInterface
}

func NewDocsTopicRepository(resourceInterface ResourceInterface) DocsTopicRepository {
	return repository{
		resourceInterface: resourceInterface,
	}
}

func (r repository) Upsert(documentationTopic docstopic.Entry) apperrors.AppError {
	docsTopic := toK8sType(documentationTopic)
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&docsTopic)
	if err != nil {
		return apperrors.Internal("Failed to convert Docs Topic object.")
	}

	u := &unstructured.Unstructured{Object: obj}

	_, err = r.resourceInterface.Update(u, metav1.UpdateOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return r.create(u)
		}

		return apperrors.Internal("Failed to create Documentation Topic")
	}

	return nil
}

func (r repository) Get(id string) (docstopic.Entry, apperrors.AppError) {
	u, err := r.resourceInterface.Get(id, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return docstopic.Entry{}, apperrors.NotFound("Docs Topic with id:%s not found", id)
		}

		if err != nil {
			return docstopic.Entry{}, nil
		}
	}

	var docsTopic v1alpha1.DocsTopic
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &docsTopic)
	if err != nil {
		return docstopic.Entry{}, apperrors.Internal("Failed to convert from unstructured object.")
	}

	return fromK8sType(docsTopic), nil
}

func (r repository) Delete(id string) apperrors.AppError {
	err := r.resourceInterface.Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return apperrors.Internal("Failed to delete DocsTopic: %s.", err)
	}

	return nil
}

func (r repository) create(u *unstructured.Unstructured) apperrors.AppError {
	_, err := r.resourceInterface.Create(u, metav1.CreateOptions{})

	if err != nil {
		return apperrors.Internal("Failed to create Documentation Topic")
	} else {
		return nil
	}
}

func toK8sType(docsTopicEntry docstopic.Entry) v1alpha1.DocsTopic {
	sources := make(map[string]v1alpha1.Source)
	for key, url := range docsTopicEntry.Urls {
		sources[key] = v1alpha1.Source{
			URL:  url,
			Mode: DocsTopicModeSingle,
		}
	}

	return v1alpha1.DocsTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DocsTopic",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      docsTopicEntry.Id,
			Namespace: "kyma-integration",
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
