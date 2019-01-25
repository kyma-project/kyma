package bucket

import (
	"context"
	"fmt"
	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/buckethandler"
	"github.com/minio/minio-go"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

var log = logf.Log.WithName("controller")

// Add creates a new Bucket Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	cfg, err := loadConfig("APP")
	if err != nil {
		return err
	}

	minioClient, err := minio.New(cfg.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.UseSSL)
	if err != nil {
		log.Error(err, "while initializing Minio client")
	}
	bucketHandler := buckethandler.New(minioClient, log)

	reconciler, err := newReconciler(mgr, bucketHandler, cfg.RequeueInterval)
	if err != nil {
		return err
	}

	return add(mgr, reconciler)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, bucketHandler buckethandler.BucketHandler, requeueInterval time.Duration) (reconcile.Reconciler, error) {
	return &ReconcileBucket{
		Client:            mgr.GetClient(),
		scheme:            mgr.GetScheme(),
		bucketHandler:     bucketHandler,
		requeueInterval:   requeueInterval,
		deletionFinalizer: &bucketFinalizer{},
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("bucket-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Bucket
	err = c.Watch(&source.Kind{Type: &assetstorev1alpha1.Bucket{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBucket{}

// ReconcileBucket reconciles a Bucket object
type ReconcileBucket struct {
	requeueInterval time.Duration
	client.Client
	scheme            *runtime.Scheme
	bucketHandler     buckethandler.BucketHandler
	deletionFinalizer *bucketFinalizer
}

// Reconcile reads that state of the cluster for a Bucket object and makes changes based on the state read
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=buckets,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileBucket) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &assetstorev1alpha1.Bucket{}
	err := r.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	// Bucket is being deleted
	handled, requeue, err := r.handleDeletionIfShould(instance)
	if handled || err != nil {
		return reconcile.Result{Requeue: requeue}, err
	}

	// Phase Empty or Failed
	if instance.Status.Phase == "" || (instance.Status.Phase == assetstorev1alpha1.BucketFailed && instance.Status.Reason != "BucketNotFound") {
		return r.handleInitialAndFailedState(instance)
	}

	// Phase Ready or Failed/BucketNotFound
	return r.handleReadyState(instance)
}

func (r *ReconcileBucket) addFinalizerIfShould(instance *assetstorev1alpha1.Bucket) {
	if r.isObjectBeingDeleted(instance) {
		return
	}

	if r.deletionFinalizer.IsDefinedIn(instance) {
		// Finalizer has been already added
		return
	}

	r.deletionFinalizer.AddTo(instance)
}

func (r *ReconcileBucket) handleDeletionIfShould(instance *assetstorev1alpha1.Bucket) (bool, bool, error) {
	if !r.isObjectBeingDeleted(instance) {
		return false, false, nil
	}

	if !r.deletionFinalizer.IsDefinedIn(instance) {
		return true, false, nil
	}

	bucketName := r.bucketNameForInstance(instance)
	err := r.bucketHandler.Delete(bucketName)
	if err != nil {
		return true, false, err
	}

	r.deletionFinalizer.DeleteFrom(instance)
	err = r.Update(context.Background(), instance)
	if err != nil {
		return true, true, nil
	}

	return true, false, nil
}

func (r *ReconcileBucket) handleInitialAndFailedState(instance *assetstorev1alpha1.Bucket) (reconcile.Result, error) {
	bucketName := r.bucketNameForInstance(instance)
	_, err := r.bucketHandler.CreateIfDoesntExist(bucketName, instance.Spec.Region)
	if err != nil {
		updateStatusErr := r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
			Phase:   assetstorev1alpha1.BucketFailed,
			Reason:  "BucketCreationFailure",
			Message: fmt.Sprintf("Bucket couldn't be created due to error %s", err.Error()),
		})
		if updateStatusErr != nil {
			return reconcile.Result{Requeue: true}, nil
		}

		return reconcile.Result{}, err
	}

	r.addFinalizerIfShould(instance)
	_ = r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
		Phase:   assetstorev1alpha1.BucketReady,
		Reason:  "BucketCreated",
		Message: "Bucket has been successfully created",
	})
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileBucket) handleReadyState(instance *assetstorev1alpha1.Bucket) (reconcile.Result, error) {
	bucketName := r.bucketNameForInstance(instance)
	policy := string(instance.Spec.Policy)

	exists, err := r.bucketHandler.CheckIfExists(bucketName)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	if !exists {
		// Bucket was created before, but has been deleted manually
		log.Info(fmt.Sprintf("Bucket %s/%s has been deleted. Setting failed status...", instance.Namespace, instance.Name))
		_ = r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
			Phase:   assetstorev1alpha1.BucketFailed,
			Reason:  "BucketNotFound",
			Message: fmt.Sprintf("Bucket %s doesn't exist anymore", bucketName),
		})

		return reconcile.Result{RequeueAfter: r.requeueInterval}, nil
	}

	// Compare policy
	updated, err := r.bucketHandler.SetPolicyIfNotEqual(bucketName, policy)
	if err != nil {
		updateStatusErr := r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
			Phase:   assetstorev1alpha1.BucketFailed,
			Reason:  "BucketPolicyUpdateFailed",
			Message: fmt.Sprintf("Bucket policy couldn't be set due to error %s", err.Error()),
		})
		if updateStatusErr != nil {
			return reconcile.Result{Requeue: true}, nil
		}

		return reconcile.Result{}, err
	}

	if updated {
		err = r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
			Phase:   assetstorev1alpha1.BucketReady,
			Reason:  "BucketPolicyUpdated",
			Message: "Bucket policy has been updated successfully",
		})
		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}

		return reconcile.Result{RequeueAfter: r.requeueInterval}, nil
	}

	// Everything is OK
	err = r.updateHeartbeatTime(instance)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	return reconcile.Result{RequeueAfter: r.requeueInterval}, nil
}

func (r *ReconcileBucket) isObjectBeingDeleted(instance *assetstorev1alpha1.Bucket) bool {
	return !instance.ObjectMeta.DeletionTimestamp.IsZero()
}

func (r *ReconcileBucket) updateStatus(instance *assetstorev1alpha1.Bucket, status assetstorev1alpha1.BucketStatus) error {
	instance.Status = status
	instance.Status.LastHeartbeatTime = metav1.Now()
	return r.Update(context.Background(), instance)
}

func (r *ReconcileBucket) updateHeartbeatTime(instance *assetstorev1alpha1.Bucket) error {
	instance.Status.LastHeartbeatTime = metav1.Now()
	return r.Update(context.Background(), instance)
}

func (r *ReconcileBucket) bucketNameForInstance(instance *assetstorev1alpha1.Bucket) string {
	return fmt.Sprintf("ns-%s-%s", instance.Namespace, instance.Name)
}
