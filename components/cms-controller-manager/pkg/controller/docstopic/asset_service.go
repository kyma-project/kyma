package docstopic

import (
	"context"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/handler/docstopic"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type assetService struct {
	client client.Client
	scheme *runtime.Scheme
}

func newAssetService(client client.Client, scheme *runtime.Scheme) *assetService {
	return &assetService{
		client: client,
		scheme: scheme,
	}
}

func (s *assetService) List(ctx context.Context, namespace string, labels map[string]string) ([]docstopic.CommonAsset, error) {
	instances := &v1alpha2.AssetList{}
	err := s.client.List(ctx, client.MatchingLabels(labels), instances)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Assets in namespace %s", namespace)
	}

	commons := make([]docstopic.CommonAsset, 0, len(instances.Items))
	for _, instance := range instances.Items {
		if instance.Namespace != namespace {
			continue
		}
		common := s.assetToCommon(instance)
		commons = append(commons, common)
	}

	return commons, nil
}

func (s *assetService) Create(ctx context.Context, docsTopic v1.Object, commonAsset docstopic.CommonAsset) error {
	instance := &v1alpha2.Asset{
		ObjectMeta: commonAsset.ObjectMeta,
		Spec: v1alpha2.AssetSpec{
			CommonAssetSpec: commonAsset.Spec,
		},
	}

	if err := controllerutil.SetControllerReference(docsTopic, instance, s.scheme); err != nil {
		return errors.Wrapf(err, "while creating Asset %s in namespace %s", commonAsset.Name, commonAsset.Namespace)
	}

	return s.client.Create(ctx, instance)
}

func (s *assetService) Update(ctx context.Context, commonAsset docstopic.CommonAsset) error {
	instance := &v1alpha2.Asset{}
	err := s.client.Get(ctx, types.NamespacedName{Name: commonAsset.Name, Namespace: commonAsset.Namespace}, instance)
	if err != nil {
		return errors.Wrapf(err, "while updating Asset %s in namespace %s", commonAsset.Name, commonAsset.Namespace)
	}

	updated := instance.DeepCopy()
	updated.Spec.CommonAssetSpec = commonAsset.Spec

	return s.client.Update(ctx, updated)
}

func (s *assetService) Delete(ctx context.Context, commonAsset docstopic.CommonAsset) error {
	instance := &v1alpha2.Asset{}
	err := s.client.Get(ctx, types.NamespacedName{Name: commonAsset.Name, Namespace: commonAsset.Namespace}, instance)
	if err != nil {
		return errors.Wrapf(err, "while deleting Asset %s in namespace %s", commonAsset.Name, commonAsset.Namespace)
	}

	return s.client.Delete(ctx, instance)
}

func (s *assetService) assetToCommon(instance v1alpha2.Asset) docstopic.CommonAsset {
	return docstopic.CommonAsset{
		ObjectMeta: instance.ObjectMeta,
		Spec:       instance.Spec.CommonAssetSpec,
		Status:     instance.Status.CommonAssetStatus,
	}
}
