package asset

import (
	"context"
	"fmt"
	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileAsset) onDelete(asset *assetstorev1alpha1.Asset, bucketReady bool) (reconcile.Result, error) {
	if !r.deleteFinalizer.IsDefinedIn(asset) {
		return reconcile.Result{}, nil
	}

	if bucketReady {
		log.Info(fmt.Sprintf("Deleting files created for asset %s/%s", asset.Namespace, asset.Name))
		bucketName := store.BucketName(asset.Namespace, asset.Spec.BucketRef.Name)
		objectPrefix := fmt.Sprintf("%s/", asset.Name)

		if err := r.cleaner.Clean(context.Background(), bucketName, objectPrefix); err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "while removing objects from store")
		}
	}

	r.deleteFinalizer.DeleteFrom(asset)
	if err := r.Update(context.Background(), asset); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "while updating Asset")
	}

	return reconcile.Result{}, nil
}
