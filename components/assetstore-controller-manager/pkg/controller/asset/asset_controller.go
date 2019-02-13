package asset

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/cleaner"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/asset/bucket"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/asset/webhook"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/finalizer"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/loader"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/uploader"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/record"
	"net/http"
	"time"

	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	uploader := uploader.New(minioClient)
	loader := loader.New(cfg.TemporaryDirectory)
	cleaner := cleaner.New(minioClient)

	deleteFinalizer := finalizer.New(deleteAssetFinalizerName)

	assethook := assethook.New(&http.Client{})
	validator := webhook.NewValidator(assethook, cfg.ValidationTimeout)
	mutator := webhook.NewMutator(assethook, cfg.MutationTimeout)

	bucketInformer, err := mgr.GetCache().GetInformer(&assetstorev1alpha1.Bucket{})
	if err != nil {
		return errors.Wrapf(err, "while initializing bucket informer")
	}
	bucketService := bucket.New(bucketInformer)

	reconciler := &ReconcileAsset{
		Client:          mgr.GetClient(),
		scheme:          mgr.GetScheme(),
		recorder:        mgr.GetRecorder("asset-controller"),
		requeueInterval: cfg.AssetRequeueInterval,

		uploader:        uploader,
		loader:          loader,
		bucketLister:    bucketService,
		cleaner:         cleaner,
		deleteFinalizer: deleteFinalizer,
		validator:       validator,
		mutator:         mutator,
	}

	return add(mgr, reconciler)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("asset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return errors.Wrapf(err, "while creating asset-controller")
	}

	// Watch for changes to Asset
	err = c.Watch(&source.Kind{Type: &assetstorev1alpha1.Asset{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return errors.Wrapf(err, "while watching Assets")
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAsset{}

// ReconcileAsset reconciles a Asset object
type ReconcileAsset struct {
	client.Client
	scheme          *runtime.Scheme
	recorder        record.EventRecorder
	requeueInterval time.Duration

	uploader        uploader.Uploader
	loader          loader.Loader
	cleaner         cleaner.Cleaner
	bucketLister    bucket.Lister
	deleteFinalizer finalizer.Finalizer
	validator       webhook.Validator
	mutator         webhook.Mutator
}

// Reconcile reads that state of the cluster for a Asset object and makes changes based on the state read
// and what is in the Asset.Spec
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=assets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=buckets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *ReconcileAsset) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &assetstorev1alpha1.Asset{}
	err := r.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, errors.Wrapf(err, "while retrieving Asset from cache")
	}

	bucket, err := r.bucketLister.Get(instance.Namespace, instance.Spec.BucketRef.Name)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "while retrieving Bucket from cache")
	}
	bucketReady := r.isBucketReady(bucket)

	switch {
	case r.isObjectBeingDeleted(&instance.ObjectMeta):
		return r.onDelete(instance, bucketReady)
	case r.isOnCreate(instance):
		return r.schedule(instance)
	case !bucketReady:
		return r.onBucketNotReady(instance)
	case instance.Status.Phase == assetstorev1alpha1.AssetReady:
		return r.onReady(instance)
	case instance.Status.Phase == assetstorev1alpha1.AssetPending:
		return r.onPending(instance, bucket)
	case r.isOnError(instance):
		return r.onError(instance)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileAsset) schedule(asset *assetstorev1alpha1.Asset) (reconcile.Result, error) {
	message := fmt.Sprintf("Scheduled asset %s/%s", asset.Namespace, asset.Name)
	r.sendEvent(asset, EventNormal, ReasonScheduled, message)
	status := r.status(assetstorev1alpha1.AssetPending, ReasonScheduled, message)

	return reconcile.Result{Requeue: true}, r.updateStatus(asset, status)
}

func (r *ReconcileAsset) isBucketReady(bucket *assetstorev1alpha1.Bucket) bool {
	return bucket != nil &&
		!r.isObjectBeingDeleted(&bucket.ObjectMeta) &&
		bucket.Status.Phase == assetstorev1alpha1.BucketReady
}

func (r *ReconcileAsset) isOnCreate(instance *assetstorev1alpha1.Asset) bool {
	return len(instance.Status.Phase) == 0
}

func (r *ReconcileAsset) isOnError(instance *assetstorev1alpha1.Asset) bool {
	return instance.Status.Phase == assetstorev1alpha1.AssetFailed &&
		instance.Status.Reason == string(ReasonError)
}

func (r *ReconcileAsset) onError(instance *assetstorev1alpha1.Asset) (reconcile.Result, error) {
	if time.Now().Before(instance.Status.LastHeartbeatTime.Time.Add(r.requeueInterval)) {
		return reconcile.Result{RequeueAfter: r.requeueInterval}, nil
	}

	return r.schedule(instance)
}

func (r *ReconcileAsset) onBucketNotReady(instance *assetstorev1alpha1.Asset) (reconcile.Result, error) {
	message := fmt.Sprintf("Bucket %s/%s for asset %s/%s is not ready", instance.Namespace, instance.Spec.BucketRef.Name, instance.Namespace, instance.Name)
	r.sendEvent(instance, EventWarning, ReasonBucketNotReady, message)

	if instance.Status.Phase == assetstorev1alpha1.AssetPending && instance.Status.Reason == string(ReasonBucketNotReady) {
		return reconcile.Result{RequeueAfter: r.requeueInterval}, nil
	}

	status := r.status(assetstorev1alpha1.AssetPending, ReasonBucketNotReady, message)
	return reconcile.Result{RequeueAfter: r.requeueInterval}, r.updateStatus(instance, status)
}

func (r *ReconcileAsset) updateStatus(instance *assetstorev1alpha1.Asset, status assetstorev1alpha1.AssetStatus) error {
	copy := instance.DeepCopy()
	copy.Status = status
	copy.Status.LastHeartbeatTime = metav1.Now()

	if err := r.Update(context.Background(), copy); err != nil {
		return errors.Wrapf(err, "while updating status")
	}

	return nil
}

func (r *ReconcileAsset) status(phase assetstorev1alpha1.AssetPhase, reason AssetReason, message string) assetstorev1alpha1.AssetStatus {
	return assetstorev1alpha1.AssetStatus{
		Phase:   phase,
		Reason:  string(reason),
		Message: message,
	}
}

func (r *ReconcileAsset) isObjectBeingDeleted(object *metav1.ObjectMeta) bool {
	return !object.DeletionTimestamp.IsZero()
}

func (r *ReconcileAsset) sendEvent(instance *assetstorev1alpha1.Asset, eventType EventLevel, reason AssetReason, message string) {
	log.Info(message)
	r.recorder.Event(instance, string(eventType), string(reason), message)
}
