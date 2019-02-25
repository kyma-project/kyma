package clusterasset

import (
	"context"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook/engine"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/finalizer"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/handler/asset"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/loader"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"time"

	assetstorev1alpha2 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
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

var log = logf.Log.WithName("clusterasset-controller")

const deleteClusterAssetFinalizerName = "deleteclusterasset.finalizers.assetstore.kyma-project.io"

// Add creates a new ClusterAsset Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
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

	store := store.New(minioClient)
	loader := loader.New(cfg.Loader.TemporaryDirectory)
	findBucketFnc := bucketFinder(mgr)
	deleteFinalizer := finalizer.New(deleteClusterAssetFinalizerName)

	assethook := assethook.New(&http.Client{})
	validator := engine.NewValidator(assethook, cfg.Webhook.ValidationTimeout)
	mutator := engine.NewMutator(assethook, cfg.Webhook.MutationTimeout)

	assetHandler := asset.New(mgr.GetRecorder("clusterasset-controller"), store, loader, findBucketFnc, validator, mutator, log)

	reconciler := &ReconcileClusterAsset{
		Client:         mgr.GetClient(),
		scheme:         mgr.GetScheme(),
		handler:        assetHandler,
		relistInterval: cfg.ClusterAssetRelistInterval,
		finalizer:      deleteFinalizer,
	}

	return add(mgr, reconciler)
}

func bucketFinder(mgr manager.Manager) func(ctx context.Context, namespace, name string) (*assetstorev1alpha2.CommonBucketStatus, bool, error) {
	return func(ctx context.Context, namespace, name string) (*assetstorev1alpha2.CommonBucketStatus, bool, error) {
		instance := &assetstorev1alpha2.ClusterBucket{}

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
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterasset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to ClusterAsset
	err = c.Watch(&source.Kind{Type: &assetstorev1alpha2.ClusterAsset{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileClusterAsset{}

// ReconcileClusterAsset reconciles a ClusterAsset object
type ReconcileClusterAsset struct {
	client.Client
	scheme *runtime.Scheme

	handler        asset.Handler
	relistInterval time.Duration
	finalizer      finalizer.Finalizer
}

// Reconcile reads that state of the cluster for a ClusterAsset object and makes changes based on the state read
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterassets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterassets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterbuckets,verbs=get;list;watch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=clusterbuckets/status,verbs=get;list
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *ReconcileClusterAsset) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	instance := &assetstorev1alpha2.ClusterAsset{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	switch {
	case !r.handler.ShouldReconcile(instance, instance.Status.CommonAssetStatus, time.Now(), time.Minute):
		return reconcile.Result{}, nil
	case r.handler.IsOnDelete(instance):
		return r.onDelete(ctx, instance)
	case r.handler.IsOnAddOrUpdate(instance, instance.Status.CommonAssetStatus):
		return r.onAddOrUpdate(ctx, instance)
	case r.handler.IsOnPending(instance.Status.CommonAssetStatus):
		return r.onPending(ctx, instance)
	case r.handler.IsOnReady(instance.Status.CommonAssetStatus):
		return r.onReady(ctx, instance)
	case r.handler.IsOnFailed(instance.Status.CommonAssetStatus):
		return r.onFailed(ctx, instance)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileClusterAsset) onDelete(ctx context.Context, instance *assetstorev1alpha2.ClusterAsset) (reconcile.Result, error) {
	if !r.finalizer.IsDefinedIn(instance) {
		return reconcile.Result{}, nil
	}

	err := r.handler.OnDelete(ctx, instance, instance.Spec.CommonAssetSpec)
	if err != nil {
		return reconcile.Result{}, err
	}

	r.finalizer.DeleteFrom(instance)
	if err := r.Update(ctx, instance); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "while updating instance")
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileClusterAsset) onPending(ctx context.Context, instance *assetstorev1alpha2.ClusterAsset) (reconcile.Result, error) {
	status := r.handler.OnPending(ctx, instance, instance.Spec.CommonAssetSpec, instance.Status.CommonAssetStatus)

	if err := r.updateStatus(ctx, instance, status); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.relistInterval}, nil
}

func (r *ReconcileClusterAsset) onReady(ctx context.Context, instance *assetstorev1alpha2.ClusterAsset) (reconcile.Result, error) {
	status := r.handler.OnReady(ctx, instance, instance.Spec.CommonAssetSpec, instance.Status.CommonAssetStatus)

	if err := r.updateStatus(ctx, instance, status); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.relistInterval}, nil
}

func (r *ReconcileClusterAsset) onFailed(ctx context.Context, instance *assetstorev1alpha2.ClusterAsset) (reconcile.Result, error) {
	status, err := r.handler.OnFailed(ctx, instance, instance.Spec.CommonAssetSpec, instance.Status.CommonAssetStatus)
	if err != nil {
		return reconcile.Result{}, err
	}

	if err := r.updateStatus(ctx, instance, *status); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.relistInterval}, nil
}

func (r *ReconcileClusterAsset) onAddOrUpdate(ctx context.Context, instance *assetstorev1alpha2.ClusterAsset) (reconcile.Result, error) {
	if !r.finalizer.IsDefinedIn(instance) {
		r.finalizer.AddTo(instance)
		return reconcile.Result{Requeue: true}, r.Update(ctx, instance)
	}
	status := r.handler.OnAddOrUpdate(ctx, instance, instance.Spec.CommonAssetSpec, instance.Status.CommonAssetStatus)

	if err := r.updateStatus(ctx, instance, status); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: r.relistInterval}, nil
}

func (r *ReconcileClusterAsset) updateStatus(ctx context.Context, instance *assetstorev1alpha2.ClusterAsset, commonStatus assetstorev1alpha2.CommonAssetStatus) error {
	toUpdate := instance.DeepCopy()
	toUpdate.Status.CommonAssetStatus = commonStatus

	if err := r.Status().Update(ctx, toUpdate); err != nil {
		return errors.Wrap(err, "while updating status")
	}

	return nil
}
