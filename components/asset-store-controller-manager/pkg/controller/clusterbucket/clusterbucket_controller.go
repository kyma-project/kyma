package clusterbucket

import (
	"context"
	assetstorev1alpha2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/finalizer"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/handler/bucket"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/store"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var log = logf.Log.WithName("clusterbucket-controller")

const deleteBucketFinalizerName = "deleteclusterbucket.finalizers.assetstore.kyma-project.io"

// Add creates a new ClusterBucket Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	cfg, err := loadConfig("APP")
	if err != nil {
		return err
	}

	minioClient, err := minio.New(cfg.Store.Endpoint, cfg.Store.AccessKey, cfg.Store.SecretKey, cfg.Store.UseSSL)
	if err != nil {
		return errors.Wrap(err, "while initializing Store client")
	}

	store := store.New(minioClient)
	deletionFinalizer := finalizer.New(deleteBucketFinalizerName)

	reconciler := &ReconcileClusterBucket{
		Client:           mgr.GetClient(),
		scheme:           mgr.GetScheme(),
		relistInterval:   cfg.ClusterBucketRelistInterval,
		finalizer:        deletionFinalizer,
		store:            store,
		externalEndpoint: cfg.Store.ExternalEndpoint,
		recorder:         mgr.GetRecorder("clusterbucket-controller"),
		cache:            mgr.GetCache(),
	}

	return add(mgr, reconciler)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterbucket-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to ClusterBucket
	err = c.Watch(&source.Kind{Type: &assetstorev1alpha2.ClusterBucket{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileClusterBucket{}

// ReconcileClusterBucket reconciles a ClusterBucket object
type ReconcileClusterBucket struct {
	client.Client
	cache    cache.Cache
	scheme   *runtime.Scheme
	recorder record.EventRecorder

	relistInterval   time.Duration
	finalizer        finalizer.Finalizer
	store            store.Store
	externalEndpoint string
}

// Reconcile reads that state of the cluster for a ClusterBucket object and makes changes based on the state read
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterbuckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterbuckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *ReconcileClusterBucket) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	if err := r.appendFinalizer(ctx, request.NamespacedName); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "while appending finalizer")
	}

	instance := &assetstorev1alpha2.ClusterBucket{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	bucketLogger := log.WithValues("kind", instance.GetObjectKind().GroupVersionKind().Kind, "name", instance.GetName())
	commonHandler := bucket.New(bucketLogger, r.recorder, r.store, r.externalEndpoint, r.relistInterval)
	commonStatus, err := commonHandler.Do(ctx, time.Now(), instance, instance.Spec.CommonBucketSpec, instance.Status.CommonBucketStatus)
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

func (r *ReconcileClusterBucket) appendFinalizer(ctx context.Context, namespacedName types.NamespacedName) error {
	updateFnc := func(instance *assetstorev1alpha2.ClusterBucket) error {
		if !instance.DeletionTimestamp.IsZero() || r.finalizer.IsDefinedIn(instance) {
			return nil
		}

		copy := instance.DeepCopy()
		r.finalizer.AddTo(copy)
		return r.Update(ctx, copy)
	}

	return r.update(ctx, namespacedName, updateFnc)
}

func (r *ReconcileClusterBucket) removeFinalizer(ctx context.Context, namespacedName types.NamespacedName) error {
	updateFnc := func(instance *assetstorev1alpha2.ClusterBucket) error {
		if instance.DeletionTimestamp.IsZero() {
			return nil
		}

		copy := instance.DeepCopy()
		r.finalizer.DeleteFrom(copy)

		return r.Update(ctx, copy)
	}

	return r.update(ctx, namespacedName, updateFnc)
}

func (r *ReconcileClusterBucket) updateStatus(ctx context.Context, namespacedName types.NamespacedName, commonStatus *assetstorev1alpha2.CommonBucketStatus) error {
	updateFnc := func(instance *assetstorev1alpha2.ClusterBucket) error {
		if r.isStatusUnchanged(instance, commonStatus) {
			return nil
		}

		copy := instance.DeepCopy()
		copy.Status.CommonBucketStatus = *commonStatus

		return r.Status().Update(ctx, copy)
	}

	return r.update(ctx, namespacedName, updateFnc)
}

func (r *ReconcileClusterBucket) isStatusUnchanged(instance *assetstorev1alpha2.ClusterBucket, newStatus *assetstorev1alpha2.CommonBucketStatus) bool {
	currentStatus := instance.Status.CommonBucketStatus

	return newStatus == nil ||
		currentStatus.ObservedGeneration == newStatus.ObservedGeneration &&
			currentStatus.Phase == newStatus.Phase &&
			currentStatus.Reason == newStatus.Reason
}

func (r *ReconcileClusterBucket) update(ctx context.Context, namespacedName types.NamespacedName, updateFnc func(instance *assetstorev1alpha2.ClusterBucket) error) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instance := &assetstorev1alpha2.ClusterBucket{}
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
