package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/assethook"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/finalizer"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/handler/asset"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/loader"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/store"
	assetstorev1alpha2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const deleteClusterAssetFinalizerName = "deleteclusterasset.finalizers.assetstore.kyma-project.io"

// ClusterAssetReconciler reconciles a ClusterAsset object
type ClusterAssetReconciler struct {
	client.Client
	Log logr.Logger

	cacheSynchronizer       func(stop <-chan struct{}) bool
	recorder                record.EventRecorder
	relistInterval          time.Duration
	maxConcurrentReconciles int
	store                   store.Store
	loader                  loader.Loader
	finalizer               finalizer.Finalizer
	validator               assethook.Validator
	mutator                 assethook.Mutator
	metadataExtractor       assethook.MetadataExtractor
}

type ClusterAssetConfig struct {
	MaxConcurrentReconciles int           `envconfig:"default=1"`
	RelistInterval          time.Duration `envconfig:"default=30s"`
}

func NewClusterAsset(config ClusterAssetConfig, log logr.Logger, di *Container) *ClusterAssetReconciler {
	deleteFinalizer := finalizer.New(deleteClusterAssetFinalizerName)

	return &ClusterAssetReconciler{
		Client:            di.Manager.GetClient(),
		cacheSynchronizer: di.Manager.GetCache().WaitForCacheSync,
		Log:               log,
		recorder:          di.Manager.GetEventRecorderFor("clusterasset-controller"),
		relistInterval:    config.RelistInterval,
		store:             di.Store,
		loader:            di.Loader,
		finalizer:         deleteFinalizer,
		validator:         di.Validator,
		mutator:           di.Mutator,
		metadataExtractor: di.Extractor,
	}
}

// Reconcile reads that state of the cluster for a ClusterAsset object and makes changes based on the state read
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterassets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterassets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterbuckets,verbs=get;list;watch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterbuckets/status,verbs=get;list
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ClusterAssetReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.appendFinalizer(ctx, request.NamespacedName); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "while appending finalizer")
	}

	instance := &assetstorev1alpha2.ClusterAsset{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	assetLogger := r.Log.WithValues("kind", instance.GetObjectKind().GroupVersionKind().Kind, "name", instance.GetName())
	commonHandler := asset.New(assetLogger, r.recorder, r.store, r.loader, r.findClusterBucket, r.validator, r.mutator, r.metadataExtractor, r.relistInterval)
	commonStatus, err := commonHandler.Do(ctx, time.Now(), instance, instance.Spec.CommonAssetSpec, instance.Status.CommonAssetStatus)
	if updateErr := r.updateStatus(ctx, request.NamespacedName, commonStatus); updateErr != nil {
		finalErr := updateErr
		if err != nil {
			finalErr = errors.Wrapf(err, "along with update error %s", updateErr.Error())
		}
		return ctrl.Result{}, finalErr
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.removeFinalizer(ctx, request.NamespacedName); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "while removing finalizer")
	}

	return ctrl.Result{
		RequeueAfter: r.relistInterval,
	}, nil
}

func (r *ClusterAssetReconciler) appendFinalizer(ctx context.Context, namespacedName types.NamespacedName) error {
	updateFnc := func(instance *assetstorev1alpha2.ClusterAsset) error {
		if !instance.DeletionTimestamp.IsZero() || r.finalizer.IsDefinedIn(instance) {
			return nil
		}

		copy := instance.DeepCopy()
		r.finalizer.AddTo(copy)
		return r.Update(ctx, copy)
	}

	return r.update(ctx, namespacedName, updateFnc)
}

func (r *ClusterAssetReconciler) removeFinalizer(ctx context.Context, namespacedName types.NamespacedName) error {
	updateFnc := func(instance *assetstorev1alpha2.ClusterAsset) error {
		if instance.DeletionTimestamp.IsZero() {
			return nil
		}

		copy := instance.DeepCopy()
		r.finalizer.DeleteFrom(copy)

		return r.Update(ctx, copy)
	}

	return r.update(ctx, namespacedName, updateFnc)
}

func (r *ClusterAssetReconciler) updateStatus(ctx context.Context, namespacedName types.NamespacedName, commonStatus *assetstorev1alpha2.CommonAssetStatus) error {
	updateFnc := func(instance *assetstorev1alpha2.ClusterAsset) error {
		if r.isStatusUnchanged(instance, commonStatus) {
			return nil
		}

		copy := instance.DeepCopy()
		copy.Status.CommonAssetStatus = *commonStatus

		return r.Status().Update(ctx, copy)
	}

	return r.update(ctx, namespacedName, updateFnc)
}

func (r *ClusterAssetReconciler) isStatusUnchanged(instance *assetstorev1alpha2.ClusterAsset, newStatus *assetstorev1alpha2.CommonAssetStatus) bool {
	currentStatus := instance.Status.CommonAssetStatus

	return newStatus == nil ||
		currentStatus.ObservedGeneration == newStatus.ObservedGeneration &&
			currentStatus.Phase == newStatus.Phase &&
			currentStatus.Reason == newStatus.Reason
}

func (r *ClusterAssetReconciler) update(ctx context.Context, namespacedName types.NamespacedName, updateFnc func(instance *assetstorev1alpha2.ClusterAsset) error) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instance := &assetstorev1alpha2.ClusterAsset{}
		err := r.Get(ctx, namespacedName, instance)
		if err != nil {
			if apiErrors.IsNotFound(err) {
				return nil
			}
			// Error reading the object - requeue the request.
			return err
		}

		err = updateFnc(instance)
		if err != nil && apiErrors.IsConflict(err) {
			r.cacheSynchronizer(ctx.Done())
		}

		return err
	})

	return err
}

func (r *ClusterAssetReconciler) findClusterBucket(ctx context.Context, namespace, name string) (*assetstorev1alpha2.CommonBucketStatus, bool, error) {
	instance := &assetstorev1alpha2.ClusterBucket{}

	namespacedName := types.NamespacedName{
		Name: name,
	}

	err := r.Get(ctx, namespacedName, instance)
	if err != nil && !apiErrors.IsNotFound(err) {
		return nil, false, err
	}

	if instance == nil || instance.Status.Phase != assetstorev1alpha2.BucketReady {
		return nil, false, nil
	}

	return &instance.Status.CommonBucketStatus, true, nil
}

func (r *ClusterAssetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&assetstorev1alpha2.ClusterAsset{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.maxConcurrentReconciles,
		}).
		Complete(r)
}
