package bucket

import (
	"context"
	"fmt"
	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/buckethandler"
	"github.com/minio/minio-go"
	"k8s.io/apimachinery/pkg/api/errors"
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
	reconciler, err := newReconciler(mgr)
	if err != nil {
		return err
	}

	return add(mgr, reconciler)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	//TODO: Load from env variables
	endpoint := "play.minio.io:9000"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true
	requeueInterval := 1 * time.Minute

	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Error(err, "while initializing Minio client")
	}
	bucketHandler := buckethandler.New(minioClient, log)

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
	bucketHandler     *buckethandler.BucketHandler
	deletionFinalizer *bucketFinalizer
}

// Reconcile reads that state of the cluster for a Bucket object and makes changes based on the state read
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=buckets/status,verbs=get;update;patch
func (r *ReconcileBucket) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &assetstorev1alpha1.Bucket{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	// TODO: Remove
	log.Info(fmt.Sprintf("Reconcile %+v", instance))
	// TODO: End

	// Bucket is being deleted
	handled, requeue, err := r.handleDeletionIfShould(instance)
	if handled || err != nil {
		return reconcile.Result{Requeue: requeue}, err
	}

	// Phase Empty or Failed
	if instance.Status.Phase == "" || instance.Status.Phase == assetstorev1alpha1.BucketFailed {
		return r.handleInitialAndFailedState(instance)
	}

	// Phase Ready
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
		_ = r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
			Phase:   assetstorev1alpha1.BucketFailed,
			Reason:  "BucketNotFound",
			Message: "Bucket doesn't exist anymore",
		})
		return reconcile.Result{Requeue: true}, nil
	}

	// Compare policy
	updated, err := r.bucketHandler.SetPolicyIfNotEqual(bucketName, policy)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	if updated {
		err = r.updateStatus(instance, assetstorev1alpha1.BucketStatus{
			Phase:   assetstorev1alpha1.BucketReady,
			Reason:  "BucketPolicyUpdated",
			Message: "Bucket policy has been updated successfully",
		})
		return reconcile.Result{Requeue: err != nil, RequeueAfter: r.requeueInterval}, nil
	}

	// Everything is OK
	return reconcile.Result{RequeueAfter: r.requeueInterval}, nil
}

func (r *ReconcileBucket) isObjectBeingDeleted(instance *assetstorev1alpha1.Bucket) bool {
	return !instance.ObjectMeta.DeletionTimestamp.IsZero()
}

func (r *ReconcileBucket) updateStatus(instance *assetstorev1alpha1.Bucket, status assetstorev1alpha1.BucketStatus) error {
	instance.Status = status
	return r.Update(context.Background(), instance)
}

func (r *ReconcileBucket) bucketNameForInstance(instance *assetstorev1alpha1.Bucket) string {
	return fmt.Sprintf("ns-%s-%s", instance.Namespace, instance.Name)
}
