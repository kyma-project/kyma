package rafter

import (
	"fmt"

	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
)

const (
	AssetGroupModeSingle = "single"
	AssetGroupNameFormat = "%s-%s"
)

//go:generate mockery -name=ResourceInterface
type ResourceInterface interface {
	Get(name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
	Delete(name string, opts *metav1.DeleteOptions, subresources ...string) error
	Create(obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error)
	Update(obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error)
}

//go:generate mockery -name=ClusterAssetGroupRepository
type ClusterAssetGroupRepository interface {
	Get(id string) (clusterassetgroup.Entry, apperrors.AppError)
	Create(assetGroup clusterassetgroup.Entry) apperrors.AppError
	Update(assetGroup clusterassetgroup.Entry) apperrors.AppError
	Delete(id string) apperrors.AppError
}

type repository struct {
	resourceInterface ResourceInterface
}

func NewAssetGroupRepository(resourceInterface ResourceInterface) ClusterAssetGroupRepository {
	return repository{
		resourceInterface: resourceInterface,
	}
}

func (r repository) Get(id string) (clusterassetgroup.Entry, apperrors.AppError) {
	assetGroup, err := r.get(id)
	if err != nil {
		return clusterassetgroup.Entry{}, err
	}

	return fromK8sType(assetGroup), nil
}

func (r repository) Delete(id string) apperrors.AppError {
	err := r.resourceInterface.Delete(id, &metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Failed to delete ClusterAssetGroup: %s.", err)
	}

	return nil
}

func (r repository) Update(assetGroupEntry clusterassetgroup.Entry) apperrors.AppError {
	k8sClusterAssetGroup := toK8sType(assetGroupEntry)
	return r.update(assetGroupEntry.Id, k8sClusterAssetGroup)
}

func (r repository) Create(assetGroupEntry clusterassetgroup.Entry) apperrors.AppError {
	k8sClusterAssetGroup := toK8sType(assetGroupEntry)
	return r.create(k8sClusterAssetGroup)
}

func (r repository) get(id string) (v1beta1.ClusterAssetGroup, apperrors.AppError) {
	u, err := r.resourceInterface.Get(id, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return v1beta1.ClusterAssetGroup{}, apperrors.NotFound("ClusterAssetGroup with %s id not found.", id)
		}

		return v1beta1.ClusterAssetGroup{}, apperrors.Internal("Failed to get ClusterAssetGroup, %s.", err)
	}

	var assetGroup v1beta1.ClusterAssetGroup
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &assetGroup)
	if err != nil {
		return v1beta1.ClusterAssetGroup{}, apperrors.Internal("Failed to convert from unstructured object, %s.", err)
	}

	return assetGroup, nil
}

func (r repository) create(assetGroup v1beta1.ClusterAssetGroup) apperrors.AppError {
	u, err := toUstructured(assetGroup)
	if err != nil {
		return apperrors.Internal("Failed to create ClusterAssetGroup, %s.", err)
	}

	_, err = r.resourceInterface.Create(u, metav1.CreateOptions{})
	if err != nil {
		return apperrors.Internal("Failed to create ClusterAssetGroup, %s.", err)
	}

	return nil
}

func (r repository) update(id string, assetGroup v1beta1.ClusterAssetGroup) apperrors.AppError {

	getRefreshedClusterAssetGroup := func(id string, assetGroup v1beta1.ClusterAssetGroup) (v1beta1.ClusterAssetGroup, error) {
		newUnstructured, err := r.resourceInterface.Get(id, metav1.GetOptions{})
		if err != nil {
			return v1beta1.ClusterAssetGroup{}, err
		}

		newClusterAssetGroup, err := fromUnstructured(newUnstructured)
		if err != nil {
			return v1beta1.ClusterAssetGroup{}, err
		}

		newClusterAssetGroup.Spec = assetGroup.Spec

		return newClusterAssetGroup, nil
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newClusterAssetGroup, err := getRefreshedClusterAssetGroup(id, assetGroup)
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
		return apperrors.Internal("Failed to update ClusterAssetGroup, %s.", err)
	}

	return nil
}

func toUstructured(assetGroup v1beta1.ClusterAssetGroup) (*unstructured.Unstructured, error) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&assetGroup)
	if err != nil {
		return nil, err
	}

	return &unstructured.Unstructured{Object: obj}, nil
}

func fromUnstructured(u *unstructured.Unstructured) (v1beta1.ClusterAssetGroup, error) {
	var assetGroup v1beta1.ClusterAssetGroup
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &assetGroup)
	if err != nil {
		return v1beta1.ClusterAssetGroup{}, err
	}

	return assetGroup, nil
}

func toK8sType(assetGroupEntry clusterassetgroup.Entry) v1beta1.ClusterAssetGroup {
	sources := make([]v1beta1.Source, 0, 3)
	annotations := make(map[string]string, len(assetGroupEntry.Assets))

	for _, asset := range assetGroupEntry.Assets {
		source := v1beta1.Source{
			Name: v1beta1.AssetGroupSourceName(fmt.Sprintf(AssetGroupNameFormat, asset.Type, asset.ID)),
			URL:  asset.Url,
			Mode: AssetGroupModeSingle,
			Type: v1beta1.AssetGroupSourceType(asset.Type),
		}

		sources = append(sources, source)
		hashAnnotationName := fmt.Sprintf(clusterassetgroup.SpecHashFormat, asset.Name)
		annotations[hashAnnotationName] = asset.SpecHash
	}

	return v1beta1.ClusterAssetGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        assetGroupEntry.Id,
			Namespace:   "kyma-integration",
			Labels:      assetGroupEntry.Labels,
			Annotations: annotations,
		},
		Spec: v1beta1.ClusterAssetGroupSpec{
			CommonAssetGroupSpec: v1beta1.CommonAssetGroupSpec{
				DisplayName: "Some display name",
				Description: "Some description",
				Sources:     sources,
			},
		}}
}

func fromK8sType(k8sAssetGroup v1beta1.ClusterAssetGroup) clusterassetgroup.Entry {
	assets := make([]clusterassetgroup.Asset, 0, len(k8sAssetGroup.Spec.Sources))

	for _, source := range k8sAssetGroup.Spec.Sources {
		asset := clusterassetgroup.Asset{
			Name: string(source.Name),
			Type: clusterassetgroup.ApiType(source.Type),
			// TODO: make sure this is needed
			Format:   "",
			Url:      source.URL,
			SpecHash: k8sAssetGroup.Annotations[fmt.Sprintf(clusterassetgroup.SpecHashFormat, source.Name)],
		}
		assets = append(assets, asset)
	}

	return clusterassetgroup.Entry{
		Id:          k8sAssetGroup.Name,
		Description: k8sAssetGroup.Spec.Description,
		DisplayName: k8sAssetGroup.Spec.DisplayName,
		Labels:      k8sAssetGroup.Labels,
		Assets:      assets,
		Status:      clusterassetgroup.StatusType(k8sAssetGroup.Status.Phase),
	}
}
