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
	Get(name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
	Delete(name string, opts *metav1.DeleteOptions, subresources ...string) error
	Create(obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error)
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

func (r repository) Upsert(docsTopicEntry docstopic.Entry) apperrors.AppError {
	existingDocsTopic, err := r.get(docsTopicEntry.Id)
	if err != nil && err.Code() == apperrors.CodeNotFound {
		return r.create(toK8sType(docsTopicEntry))
	}

	if err != nil {
		return err
	}

	k8sDocsTopic := toK8sType(docsTopicEntry)
	k8sDocsTopic.ResourceVersion = existingDocsTopic.ResourceVersion

	return r.update(k8sDocsTopic)
}

func (r repository) Get(id string) (docstopic.Entry, apperrors.AppError) {
	docsTopic, err := r.get(id)
	if err != nil {
		return docstopic.Entry{}, err
	}

	return fromK8sType(docsTopic), nil
}

func (r repository) Delete(id string) apperrors.AppError {
	err := r.resourceInterface.Delete(id, &metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Failed to delete DocsTopic: %s.", err)
	}

	return nil
}

func (r repository) get(id string) (v1alpha1.DocsTopic, apperrors.AppError) {
	u, err := r.resourceInterface.Get(id, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return v1alpha1.DocsTopic{}, apperrors.NotFound("Docs Topic with %s id not found.", id)
		}

		return v1alpha1.DocsTopic{}, apperrors.Internal("Failed to get Docs Topic, %s.", err)
	}

	var docsTopic v1alpha1.DocsTopic
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &docsTopic)
	if err != nil {
		return v1alpha1.DocsTopic{}, apperrors.Internal("Failed to convert from unstructured object, %s.", err)
	}

	return docsTopic, nil
}

func (r repository) create(docsTopic v1alpha1.DocsTopic) apperrors.AppError {
	u, err := toUstructured(docsTopic)
	if err != nil {
		return err
	}

	{
		_, err := r.resourceInterface.Create(u, metav1.CreateOptions{})

		if err != nil {
			return apperrors.Internal("Failed to create Documentation Topic, %s.", err)
		}
	}

	return nil
}

func (r repository) update(docsTopic v1alpha1.DocsTopic) apperrors.AppError {
	u, err := toUstructured(docsTopic)
	if err != nil {
		return err
	}

	{
		_, err := r.resourceInterface.Update(u, metav1.UpdateOptions{})

		if err != nil {
			return apperrors.Internal("Failed to update Documentation Topic, %s.", err)
		}
	}

	return nil
}

func toUstructured(docsTopic v1alpha1.DocsTopic) (*unstructured.Unstructured, apperrors.AppError) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&docsTopic)
	if err != nil {
		return nil, apperrors.Internal("Failed to convert Docs Topic object, %s.", err)
	}

	return &unstructured.Unstructured{Object: obj}, nil
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
		Status:      docstopic.StatusType(k8sDocsTopic.Status.Phase),
	}
}
