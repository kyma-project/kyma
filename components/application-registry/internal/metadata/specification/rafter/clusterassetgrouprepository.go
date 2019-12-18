package rafter

import (
	"fmt"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/clusterassetgroup"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
)

const (
	ClusterAssetGroupModeSingle = "single"
	ClusterAssetGroupNameFormat = "%s-%s"
)

type ResourceInterface interface {
	Get(name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
	Delete(name string, opts *metav1.DeleteOptions, subresources ...string) error
	Create(obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error)
	Update(obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error)
}

type ClusterAssetGroupRepository interface {
	Get(id string) (clusterassetgroup.Entry, apperrors.AppError)
	Upsert(documentationTopic clusterassetgroup.Entry) apperrors.AppError
	Delete(id string) apperrors.AppError
}

type repository struct {
	resourceInterface ResourceInterface
}

func NewClusterAssetGroupRepository(resourceInterface ResourceInterface) ClusterAssetGroupRepository {
	return repository{
		resourceInterface: resourceInterface,
	}
}

func (r repository) Upsert(clusterAssetGroupEntry clusterassetgroup.Entry) apperrors.AppError {
	_, err := r.get(clusterAssetGroupEntry.Id)
	if err != nil && err.Code() == apperrors.CodeNotFound {
		return r.create(toK8sType(clusterAssetGroupEntry))
	}

	if err != nil {
		return err
	}

	k8sClusterAssetGroup := toK8sType(clusterAssetGroupEntry)

	return r.update(clusterAssetGroupEntry.Id, k8sClusterAssetGroup)
}

func (r repository) Get(id string) (clusterassetgroup.Entry, apperrors.AppError) {
	clusterAssetGroup, err := r.get(id)
	if err != nil {
		return clusterassetgroup.Entry{}, err
	}

	return fromK8sType(clusterAssetGroup), nil
}

func (r repository) Delete(id string) apperrors.AppError {
	err := r.resourceInterface.Delete(id, &metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Failed to delete ClusterAssetGroup: %s.", err)
	}

	return nil
}

func (r repository) get(id string) (v1beta1.ClusterAssetGroup, apperrors.AppError) {
	u, err := r.resourceInterface.Get(id, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return v1beta1.ClusterAssetGroup{}, apperrors.NotFound("Docs Topic with %s id not found.", id)
		}

		return v1beta1.ClusterAssetGroup{}, apperrors.Internal("Failed to get Docs Topic, %s.", err)
	}

	var clusterAssetGroup v1beta1.ClusterAssetGroup
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &clusterAssetGroup)
	if err != nil {
		return v1beta1.ClusterAssetGroup{}, apperrors.Internal("Failed to convert from unstructured object, %s.", err)
	}

	return clusterAssetGroup, nil
}

func (r repository) create(clusterAssetGroup v1beta1.ClusterAssetGroup) apperrors.AppError {
	u, err := toUstructured(clusterAssetGroup)
	if err != nil {
		return apperrors.Internal("Failed to create Cluster Asset Group: %s.", err)
	}

	_, err = r.resourceInterface.Create(u, metav1.CreateOptions{})
	if err != nil {
		return apperrors.Internal("Failed to create Cluster Asset Group: %s.", err)
	}

	return nil
}

func (r repository) update(id string, clusterAssetGroup v1beta1.ClusterAssetGroup) apperrors.AppError {

	getRefreshedClusterAssetGroup := func(id string, clusterAssetGroup v1beta1.ClusterAssetGroup) (v1beta1.ClusterAssetGroup, error) {
		newUnstructured, err := r.resourceInterface.Get(id, metav1.GetOptions{})
		if err != nil {
			return v1beta1.ClusterAssetGroup{}, err
		}

		newClusterAssetGroup, err := fromUnstructured(newUnstructured)
		if err != nil {
			return v1beta1.ClusterAssetGroup{}, err
		}

		newClusterAssetGroup.Spec = clusterAssetGroup.Spec

		return newClusterAssetGroup, nil
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newClusterAssetGroup, err := getRefreshedClusterAssetGroup(id, clusterAssetGroup)
		if err != nil {
			return err
		}

		u, err := toUstructured(newClusterAssetGroup)
		if err != nil {
			return err
		}

		_, err = r.resourceInterface.Update(u, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return apperrors.Internal("Failed to update Cluster Asset Group: %s.", err)
	}

	return nil
}

func toUstructured(clusterAssetGroup v1beta1.ClusterAssetGroup) (*unstructured.Unstructured, error) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&clusterAssetGroup)
	if err != nil {
		return nil, err
	}

	return &unstructured.Unstructured{Object: obj}, nil
}

func fromUnstructured(u *unstructured.Unstructured) (v1beta1.ClusterAssetGroup, error) {
	var clusterAssetGroup v1beta1.ClusterAssetGroup
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &clusterAssetGroup)
	if err != nil {
		return v1beta1.ClusterAssetGroup{}, err
	}

	return clusterAssetGroup, nil
}

func toK8sType(clusterAssetGroup clusterassetgroup.Entry) v1beta1.ClusterAssetGroup {
	sources := make([]v1beta1.Source, 0, 3)
	for key, url := range clusterAssetGroup.Urls {
		source := v1beta1.Source{
			Name: v1beta1.AssetGroupSourceName(fmt.Sprintf(ClusterAssetGroupNameFormat, key, clusterAssetGroup.Id)),
			URL:  url,
			Mode: ClusterAssetGroupModeSingle,
			Type: v1beta1.AssetGroupSourceType(key),
		}
		sources = append(sources, source)
	}

	return v1beta1.ClusterAssetGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterAssetGroup.Id,
			Namespace: "kyma-integration",
			Labels:    clusterAssetGroup.Labels,
		},
		Spec: v1beta1.ClusterAssetGroupSpec{
			CommonAssetGroupSpec: v1beta1.CommonAssetGroupSpec{
				DisplayName: "Some display name",
				Description: "Some description",
				Sources:     sources,
			},
		}}
}

func fromK8sType(k8sClusterAssetGroup v1beta1.ClusterAssetGroup) clusterassetgroup.Entry {
	urls := make(map[string]string)

	for _, source := range k8sClusterAssetGroup.Spec.Sources {
		urls[string(source.Type)] = source.URL
	}

	return clusterassetgroup.Entry{
		Id:          k8sClusterAssetGroup.Name,
		Description: k8sClusterAssetGroup.Spec.Description,
		DisplayName: k8sClusterAssetGroup.Spec.DisplayName,
		Urls:        urls,
		Labels:      k8sClusterAssetGroup.Labels,
		Status:      clusterassetgroup.StatusType(k8sClusterAssetGroup.Status.Phase),
	}
}
