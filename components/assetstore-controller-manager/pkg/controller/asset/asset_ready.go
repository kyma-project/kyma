package asset

import (
	"fmt"
	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileAsset) onReady(asset *assetstorev1alpha1.Asset) (reconcile.Result, error) {
	log.Info(fmt.Sprintf("Checking if files created by asset %s/%s exists", asset.Namespace, asset.Name))
	bucketName := store.BucketName(asset.Namespace, asset.Spec.BucketRef.Name)
	contains, err := r.uploader.ContainsAll(bucketName, asset.Name, asset.Status.AssetRef.Assets)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "while checking files in bucket")
	}

	if !contains {
		return reconcile.Result{}, r.setStatusPendingMissing(asset)
	}

	return reconcile.Result{RequeueAfter: r.requeueInterval}, nil
}

func (r *ReconcileAsset) setStatusPendingMissing(instance *assetstorev1alpha1.Asset) error {
	message := fmt.Sprintf("Missing files for asset %s/%s", instance.Namespace, instance.Name)
	r.sendEvent(instance, EventWarning, ReasonMissingFiles, message)

	status := r.status(assetstorev1alpha1.AssetPending, ReasonMissingFiles, message)
	r.deleteFinalizer.DeleteFrom(instance)

	return r.updateStatus(instance, status)
}
