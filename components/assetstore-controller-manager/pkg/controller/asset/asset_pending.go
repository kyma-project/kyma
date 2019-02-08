package asset

import (
	"context"
	"fmt"
	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/asset/webhook"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileAsset) onPending(asset *assetstorev1alpha1.Asset, bucket *assetstorev1alpha1.Bucket) (reconcile.Result, error) {
	basePath, files, err := r.download(asset)
	defer r.loader.Clean(basePath)
	if err != nil {
		return r.setStatusFailed(asset, ReasonError, fmt.Sprintf("Cannot download asset: %s", err.Error()))
	}

	result, err := r.validate(context.Background(), asset, basePath, files)
	if err != nil {
		return r.setStatusFailed(asset, ReasonError, fmt.Sprintf("Cannot validate asset: %s", err.Error()))
	}
	if !result.Success {
		return r.setStatusFailed(asset, ReasonValidationFailed, fmt.Sprintf("Validation failed: %+v", result.Messages))
	}

	if err := r.mutate(context.Background(), asset, basePath, files); err != nil {
		return r.setStatusFailed(asset, ReasonMutationFailed, fmt.Sprintf("Cannot mutate asset: %s", err.Error()))
	}

	if err := r.upload(context.Background(), asset, basePath, files); err != nil {
		return r.setStatusFailed(asset, ReasonError, fmt.Sprintf("Cannot upload asset to store: %s", err.Error()))
	}
	r.sendEvent(asset, EventNormal, ReasonUploaded, "Uploaded files to bucket")

	if err := r.setStatusReady(asset, bucket.Status.Url, files); err != nil {
		log.Info(fmt.Sprintf("Error while setting status to Ready: %+v", err))
		return reconcile.Result{RequeueAfter: r.requeueInterval}, err
	}

	return reconcile.Result{RequeueAfter: r.requeueInterval}, nil
}

func (r *ReconcileAsset) setStatusFailed(instance *assetstorev1alpha1.Asset, reason AssetReason, message string) (reconcile.Result, error) {
	r.sendEvent(instance, EventWarning, reason, message)
	status := r.status(assetstorev1alpha1.AssetFailed, reason, message)
	return reconcile.Result{RequeueAfter: r.requeueInterval}, r.updateStatus(instance, status)
}

func (r *ReconcileAsset) validate(ctx context.Context, instance *assetstorev1alpha1.Asset, basePath string, files []string) (webhook.ValidationResult, error) {
	if len(instance.Spec.Source.ValidationWebhookService) == 0 {
		return webhook.ValidationResult{Success: true}, nil
	}

	result, err := r.validator.Validate(ctx, basePath, files, instance)
	if err != nil {
		return webhook.ValidationResult{}, err
	}

	r.sendEvent(instance, EventNormal, ReasonValidated, "Validated files provided by asset")

	return result, nil
}

func (r *ReconcileAsset) mutate(ctx context.Context, instance *assetstorev1alpha1.Asset, basePath string, files []string) error {
	if len(instance.Spec.Source.MutationWebhookService) == 0 {
		return nil
	}

	if err := r.mutator.Mutate(ctx, basePath, files, instance); err != nil {
		return err
	}

	r.sendEvent(instance, EventNormal, ReasonMutated, "Mutated files provided by asset")

	return nil
}

func (r *ReconcileAsset) upload(ctx context.Context, instance *assetstorev1alpha1.Asset, basePath string, files []string) error {
	bucketName := store.BucketName(instance.Namespace, instance.Spec.BucketRef.Name)

	return r.uploader.Upload(ctx, bucketName, instance.Name, basePath, files)
}

func (r *ReconcileAsset) download(instance *assetstorev1alpha1.Asset) (string, []string, error) {
	basePath, files, err := r.loader.Load(instance.Spec.Source.Url, instance.Name, instance.Spec.Source.Mode)
	if err != nil {
		return "", nil, errors.Wrapf(err, "while downloading Asset")
	}
	r.sendEvent(instance, EventNormal, ReasonPulled, "Pulled files")

	return basePath, files, nil
}

func (r *ReconcileAsset) setStatusReady(instance *assetstorev1alpha1.Asset, bucketUrl string, assets []string) error {
	status := r.status(assetstorev1alpha1.AssetReady, ReasonReady, "Asset is ready to use")
	status.AssetRef = assetstorev1alpha1.AssetStatusRef{
		BaseUrl: fmt.Sprintf("%s/%s", bucketUrl, instance.Name),
		Assets:  assets,
	}
	r.deleteFinalizer.AddTo(instance)

	return r.updateStatus(instance, status)
}
