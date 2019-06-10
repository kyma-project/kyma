package asset

import (
	"context"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/engine"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/finalizer"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/handler/asset"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/loader"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/store"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"time"

	assetstorev1alpha2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("asset-controller")

const deleteAssetFinalizerName = "deleteasset.finalizers.assetstore.kyma-project.io"

// Add creates a new Asset Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	cfg, err := loadConfig("APP")
	if err != nil {
		return errors.Wrapf(err, "while loading configuration")
	}

	minioClient, err := minio.New(cfg.Store.Endpoint, cfg.Store.AccessKey, cfg.Store.SecretKey, cfg.Store.UseSSL)
	if err != nil {
		return errors.Wrapf(err, "while initializing Store client")
	}

	store := store.New(minioClient, cfg.Store.UploadWorkerCount)
	loader := loader.New(cfg.Loader.TemporaryDirectory, cfg.Loader.VerifySSL)
	findBucketFnc := bucketFinder(mgr)
	deleteFinalizer := finalizer.New(deleteAssetFinalizerName)

	assethook := assethook.New(&http.Client{})
	validator := engine.NewValidator(assethook, cfg.Webhook.ValidationTimeout)
	mutator := engine.NewMutator(assethook, cfg.Webhook.MutationTimeout)
	metadataExtractor := engine.NewMetadataExtractor(assethook, cfg.Webhook.MetadataExtractionTimeout)

	reconciler := &ReconcileAsset{
		Client:            mgr.GetClient(),
		cache:             mgr.GetCache(),
		scheme:            mgr.GetScheme(),
		store:             store,
		loader:            loader,
		findBucketFnc:     findBucketFnc,
		validator:         validator,
		mutator:           mutator,
		recorder:          mgr.GetRecorder("asset-controller"),
		relistInterval:    cfg.AssetRelistInterval,
		finalizer:         deleteFinalizer,
		metadataExtractor: metadataExtractor,
	}

	return add(mgr, reconciler, cfg.MaxAssetConcurrentReconciles)
}

func bucketFinder(mgr manager.Manager) func(ctx context.Context, namespace, name string) (*assetstorev1alpha2.CommonBucketStatus, bool, error) {
	return func(ctx context.Context, namespace, name string) (*assetstorev1alpha2.CommonBucketStatus, bool, error) {
		instance := &assetstorev1alpha2.Bucket{}

		namespacedName := types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}

		err := mgr.GetCache().Get(ctx, namespacedName, instance)
		if err != nil && !apiErrors.IsNotFound(err) {
			return nil, false, err
		}

		if instance == nil || instance.Status.Phase != assetstorev1alpha2.BucketReady {
			return nil, false, nil
		}

		return &instance.Status.CommonBucketStatus, true, nil
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, maxConcurrentReconciles int) error {
	// Create a new controller
	c, err := controller.New("asset-controller", mgr, controller.Options{
		Reconciler:              r,
		MaxConcurrentReconciles: maxConcurrentReconciles,
	})
	if err != nil {
		return errors.Wrapf(err, "while creating asset-controller")
	}

	// Watch for changes to Asset
	err = c.Watch(&source.Kind{Type: &assetstorev1alpha2.Asset{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return errors.Wrapf(err, "while watching Assets")
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAsset{}

// ReconcileAsset reconciles a Asset object
type ReconcileAsset struct {
	client.Client
	cache    cache.Cache
	scheme   *runtime.Scheme
	recorder record.EventRecorder

	relistInterval    time.Duration
	store             store.Store
	loader            loader.Loader
	findBucketFnc     asset.FindBucketStatus
	finalizer         finalizer.Finalizer
	validator         engine.Validator
	mutator           engine.Mutator
	metadataExtractor engine.MetadataExtractor
}

// Reconcile reads that state of the cluster for a Asset object and makes changes based on the state read
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=assets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=assets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=buckets,verbs=get;list;watch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=buckets/status,verbs=get;list
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *ReconcileAsset) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	if err := r.appendFinalizer(ctx, request.NamespacedName); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "while appending finalizer")
	}

	instance := &assetstorev1alpha2.Asset{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	assetLogger := log.WithValues("kind", instance.GetObjectKind().GroupVersionKind().Kind, "name", instance.GetName(), "namespace", instance.GetNamespace())
	commonHandler := asset.New(assetLogger, r.recorder, r.store, r.loader, r.findBucketFnc, r.validator, r.mutator, r.metadataExtractor, r.relistInterval)
	commonStatus, err := commonHandler.Do(ctx, time.Now(), instance, instance.Spec.CommonAssetSpec, instance.Status.CommonAssetStatus)
	if updateErr := r.updateStatus(ctx, request.NamespacedName, commonStatus); updateErr != nil {
		finalErr := updateErr
		if err != nil {
			finalErr = errors.Wrapf(err, "along with update error %s", updateErr.Error())
		}
		return reconcile.Result{}, finalErr
	}
	if err != nil {
		return reconcile.Result{}, err
	}

	if err := r.removeFinalizer(ctx, request.NamespacedName); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "while removing finalizer")
	}

	return reconcile.Result{
		RequeueAfter: r.relistInterval,
	}, nil
}

func (r *ReconcileAsset) appendFinalizer(ctx context.Context, namespacedName types.NamespacedName) error {
	updateFnc := func(instance *assetstorev1alpha2.Asset) error {
		if !instance.DeletionTimestamp.IsZero() || r.finalizer.IsDefinedIn(instance) {
			return nil
		}

		copy := instance.DeepCopy()
		r.finalizer.AddTo(copy)
		return r.Update(ctx, copy)
	}

	return r.update(ctx, namespacedName, updateFnc)
}

func (r *ReconcileAsset) removeFinalizer(ctx context.Context, namespacedName types.NamespacedName) error {
	updateFnc := func(instance *assetstorev1alpha2.Asset) error {
		if instance.DeletionTimestamp.IsZero() {
			return nil
		}

		copy := instance.DeepCopy()
		r.finalizer.DeleteFrom(copy)

		return r.Update(ctx, copy)
	}

	return r.update(ctx, namespacedName, updateFnc)
}

func (r *ReconcileAsset) updateStatus(ctx context.Context, namespacedName types.NamespacedName, commonStatus *assetstorev1alpha2.CommonAssetStatus) error {
	updateFnc := func(instance *assetstorev1alpha2.Asset) error {
		if r.isStatusUnchanged(instance, commonStatus) {
			return nil
		}

		copy := instance.DeepCopy()
		copy.Status.CommonAssetStatus = *commonStatus

		return r.Status().Update(ctx, copy)
	}

	return r.update(ctx, namespacedName, updateFnc)
}

func (r *ReconcileAsset) isStatusUnchanged(instance *assetstorev1alpha2.Asset, newStatus *assetstorev1alpha2.CommonAssetStatus) bool {
	currentStatus := instance.Status.CommonAssetStatus

	return newStatus == nil ||
		currentStatus.ObservedGeneration == newStatus.ObservedGeneration &&
			currentStatus.Phase == newStatus.Phase &&
			currentStatus.Reason == newStatus.Reason
}

func (r *ReconcileAsset) update(ctx context.Context, namespacedName types.NamespacedName, updateFnc func(instance *assetstorev1alpha2.Asset) error) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instance := &assetstorev1alpha2.Asset{}
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
			r.cache.WaitForCacheSync(ctx.Done())
		}

		return err
	})

	return err
}
