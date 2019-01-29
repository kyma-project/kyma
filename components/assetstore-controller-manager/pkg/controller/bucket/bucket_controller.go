package bucket

import (
	"context"
	"fmt"
	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/buckethandler"
	"github.com/minio/minio-go"
	pkgErrors "github.com/pkg/errors"
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
		return pkgErrors.Wrap(err, "while initializing Minio client")
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
	handled, err := r.handleDeletionIfShould(instance)
	if handled || err != nil {
		return reconcile.Result{}, err
	}

	// Phase Empty or Failed
	if instance.Status.Phase == "" || (instance.Status.Phase == assetstorev1alpha1.BucketFailed && instance.Status.Reason != ReasonNotFound.String()) {
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

func (r *ReconcileBucket) handleDeletionIfShould(instance *assetstorev1alpha1.Bucket) (bool, error) {
	if !r.isObjectBeingDeleted(instance) {
		return false, nil
	}

	if !r.deletionFinalizer.IsDefinedIn(instance) {
		return true, nil
	}

	bucketName := r.bucketNameForInstance(instance)
	err := r.bucketHandler.Delete(bucketName)
	if err != nil {
		return false, err
	}

	r.deletionFinalizer.DeleteFrom(instance)
	err = r.Update(context.Background(), instance)
	if err != nil {
		return true, err
	}

	return true, nil
}

func (r *ReconcileBucket) handleInitialAndFailedState(instance *assetstorev1alpha1.Bucket) (reconcile.Result, error) {
	bucketName := r.bucketNameForInstance(instance)
	handled, err := r.bucketHandler.CreateIfDoesntExist(bucketName, string(instance.Spec.Region))
	if err != nil {
		updateStatusErr := r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
			Phase:   assetstorev1alpha1.BucketFailed,
			Reason:  BucketCreationFailure.String(),
			Message: fmt.Sprintf("Bucket couldn't be created due to error %s", err.Error()),
		})
		if updateStatusErr != nil {
			return reconcile.Result{}, updateStatusErr
		}

		return reconcile.Result{}, err
	}

	if !handled && instance.Status.Reason == ReasonPolicyUpdateFailed.String() {
		return r.updatePolicyIfShould(instance)
	}

	r.addFinalizerIfShould(instance)
	err = r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
		Phase:   assetstorev1alpha1.BucketReady,
		Reason:  BucketCreated.String(),
		Message: "Bucket has been successfully created",
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileBucket) handleReadyState(instance *assetstorev1alpha1.Bucket) (reconcile.Result, error) {
	bucketName := r.bucketNameForInstance(instance)

	exists, err := r.bucketHandler.Exists(bucketName)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !exists && instance.Status.Reason == string(ReasonNotFound) {
		log.Info(fmt.Sprintf("Bucket %s still doesn't exist. Scheduling requeue...", bucketName))
		return reconcile.Result{RequeueAfter: r.requeueInterval}, nil
	}

	if !exists {
		// Bucket was created before, but has been deleted manually
		log.Info(fmt.Sprintf("bucket %s/%s has been deleted", instance.Namespace, instance.Name))
		r.deletionFinalizer.DeleteFrom(instance)
		_ = r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
			Phase:   assetstorev1alpha1.BucketFailed,
			Reason:  ReasonNotFound.String(),
			Message: fmt.Sprintf("Bucket %s doesn't exist anymore", bucketName),
		})
		return reconcile.Result{}, nil
	}

	// Compare policy
	return r.updatePolicyIfShould(instance)
}

func (r *ReconcileBucket) updatePolicyIfShould(instance *assetstorev1alpha1.Bucket) (reconcile.Result, error) {
	bucketName := r.bucketNameForInstance(instance)
	policy := string(instance.Spec.Policy)

	// Compare policy
	updated, err := r.bucketHandler.SetPolicyIfNotEqual(bucketName, policy)
	if err != nil {
		updateStatusErr := r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
			Phase:   assetstorev1alpha1.BucketFailed,
			Reason:  BucketPolicyUpdateFailed.String(),
			Message: fmt.Sprintf("Bucket policy couldn't be set due to error %s", err.Error()),
		})
		if updateStatusErr != nil {
			return reconcile.Result{}, updateStatusErr
		}

		return reconcile.Result{}, err
	}

	if updated {
		err = r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
			Phase:   assetstorev1alpha1.BucketReady,
			Reason:  BucketPolicyUpdated.String(),
			Message: "Bucket policy has been updated successfully",
		})
		if err != nil {
			return reconcile.Result{}, err
		}
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

func (r *ReconcileBucket) bucketNameForInstance(instance *assetstorev1alpha1.Bucket) string {
	return fmt.Sprintf("ns-%s-%s", instance.Namespace, instance.Name)
}
