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
		return reconcile.Result{}, err
	}

	result, err := r.validate(context.Background(), asset, basePath, files)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "while validating Asset")
	}
	if !result.Success {
		return reconcile.Result{}, r.setStatusWebhookFailed(asset, ReasonValidationFailed, fmt.Sprintf("%+v", result.Messages))
	}

	if err := r.mutate(context.Background(), asset, basePath, files); err != nil {
		r.setStatusWebhookFailed(asset, ReasonMutationFailed, fmt.Sprintf("%+v", err))
		return reconcile.Result{}, errors.Wrapf(err, "while mutating Asset")
	}

	if err := r.upload(context.Background(), asset, basePath, files); err != nil {
		r.sendEvent(asset, EventWarning, ReasonError, "Upload to bucket failed")
		return reconcile.Result{}, errors.Wrapf(err, "while uploading Asset")
	}
	r.sendEvent(asset, EventNormal, ReasonUploaded, "Uploaded files to bucket")

	if err := r.setStatusReady(asset, bucket.Status.Url, files); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.requeueInterval}, nil
}

func (r *ReconcileAsset) setStatusWebhookFailed(instance *assetstorev1alpha1.Asset, reason AssetReason, message string) error {
	r.sendEvent(instance, EventWarning, reason, message)
	status := r.status(assetstorev1alpha1.AssetFailed, reason, message)

	return r.updateStatus(instance, status)
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
