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
	handler := bucket.New(mgr.GetRecorder("clusterbucket-controller"), store, cfg.Store.ExternalEndpoint, log)

	reconciler := &ReconcileClusterBucket{
		Client:         mgr.GetClient(),
		scheme:         mgr.GetScheme(),
		handler:        handler,
		relistInterval: cfg.ClusterBucketRelistInterval,
		finalizer:      deletionFinalizer,
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
	scheme *runtime.Scheme

	handler        bucket.Handler
	relistInterval time.Duration
	finalizer      finalizer.Finalizer
}

// Reconcile reads that state of the cluster for a ClusterBucket object and makes changes based on the state read
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterbuckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterbuckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *ReconcileClusterBucket) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	instance := &assetstorev1alpha2.ClusterBucket{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	switch {
	case !r.handler.ShouldReconcile(instance, instance.Status.CommonBucketStatus, time.Now(), r.relistInterval):
		return reconcile.Result{}, nil
	case r.handler.IsOnDelete(instance):
		return r.onDelete(ctx, instance)
	case r.handler.IsOnAddOrUpdate(instance, instance.Status.CommonBucketStatus):
		return r.onAddOrUpdate(ctx, instance)
	case r.handler.IsOnReady(instance.Status.CommonBucketStatus):
		return r.onReady(ctx, instance)
	case r.handler.IsOnFailed(instance.Status.CommonBucketStatus):
		return r.onFailed(ctx, instance)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileClusterBucket) onFailed(ctx context.Context, instance *assetstorev1alpha2.ClusterBucket) (reconcile.Result, error) {
	status, err := r.handler.OnFailed(instance, instance.Spec.CommonBucketSpec, instance.Status.CommonBucketStatus)
	if err != nil {
		return reconcile.Result{}, err
	}

	if err := r.updateStatus(ctx, instance, *status); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.relistInterval}, nil
}

func (r *ReconcileClusterBucket) onReady(ctx context.Context, instance *assetstorev1alpha2.ClusterBucket) (reconcile.Result, error) {
	status := r.handler.OnReady(instance, instance.Spec.CommonBucketSpec, instance.Status.CommonBucketStatus)

	if err := r.updateStatus(ctx, instance, status); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.relistInterval}, nil
}

func (r *ReconcileClusterBucket) onDelete(ctx context.Context, instance *assetstorev1alpha2.ClusterBucket) (reconcile.Result, error) {
	if !r.finalizer.IsDefinedIn(instance) {
		return reconcile.Result{}, nil
	}

	err := r.handler.OnDelete(ctx, instance, instance.Status.CommonBucketStatus)
	if err != nil {
		return reconcile.Result{}, err
	}

	r.finalizer.DeleteFrom(instance)

	if err := r.Update(ctx, instance); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "while updating instance")
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileClusterBucket) onAddOrUpdate(ctx context.Context, instance *assetstorev1alpha2.ClusterBucket) (reconcile.Result, error) {
	if !r.finalizer.IsDefinedIn(instance) {
		r.finalizer.AddTo(instance)
		return reconcile.Result{Requeue: true}, r.Update(ctx, instance)
	}

	status := r.handler.OnAddOrUpdate(instance, instance.Spec.CommonBucketSpec, instance.Status.CommonBucketStatus)

	if err := r.updateStatus(ctx, instance, status); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.relistInterval}, nil
}

func (r *ReconcileClusterBucket) updateStatus(ctx context.Context, instance *assetstorev1alpha2.ClusterBucket, commonStatus assetstorev1alpha2.CommonBucketStatus) error {
	instance.Status.CommonBucketStatus = commonStatus

	if err := r.Status().Update(ctx, instance); err != nil {
		return errors.Wrap(err, "while updating status")
	}

	return nil
}
