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
		Client:          mgr.GetClient(),
		scheme:          mgr.GetScheme(),
		bucketHandler:   bucketHandler,
		requeueInterval: requeueInterval,
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
	scheme        *runtime.Scheme
	bucketHandler *buckethandler.BucketHandler
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

	bucketName := r.bucketNameForInstance(instance)

	// TODO: Remove
	log.Info(fmt.Sprintf("Reconcile %+v", instance))

	// name of your custom finalizer
	deleteBucketFinalizerName := "deletebucket.finalizers.assetstore.kyma-project.io"

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !containsString(instance.ObjectMeta.Finalizers, deleteBucketFinalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, deleteBucketFinalizerName)
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{Requeue: true}, nil
			}
		}
	} else {
		// The object is being deleted
		if containsString(instance.ObjectMeta.Finalizers, deleteBucketFinalizerName) {
			// our finalizer is present, so lets handle minio bucket deletion
			if err := r.bucketHandler.Delete(instance); err != nil {
				// if fail to delete bucket here, return with error
				// so that it can be retried
				return reconcile.Result{}, err
			}

			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, deleteBucketFinalizerName)
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{Requeue: true}, nil
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return reconcile.Result{}, nil
	}

	bucketName := r.bucketNameForInstance(instance)
	exists, err := r.minioClient.BucketExists(bucketName)

	if exists {
		instance.Status.Phase = assetstorev1alpha1.BucketReady
		instance.Status.Reason = "Created"
		instance.Status.Message = "Bucket %s successfully created"
		updateErr := r.Update(context.Background(), instance)
		if updateErr != nil {
			return reconcile.Result{Requeue: true}, nil
		}

		return reconcile.Result{RequeueAfter: 10 * time.Minute}, nil
	}

	if instance.Status.Phase == "" {
		//Empty status - make bucket

		bucketName := r.bucketNameForInstance(instance)

		err = r.minioClient.MakeBucket(bucketName, region)
		if err != nil {
			// Bucket failed to create

			instance.Status.Phase = assetstorev1alpha1.BucketFailed
			instance.Status.Reason = "CreationFailed"
			instance.Status.Message = fmt.Sprintf("Bucket couldn't be created due to error %s", err)
			updateErr := r.Update(context.Background(), instance)
			if updateErr != nil {
				return reconcile.Result{Requeue: true}, nil
			}

			return reconcile.Result{}, err
		}

		instance.Status.Phase = assetstorev1alpha1.BucketReady
		instance.Status.Reason = "Created"
		instance.Status.Message = "Bucket %s successfully created"
		updateErr := r.Update(context.Background(), instance)
		if updateErr != nil {
			return reconcile.Result{Requeue: true}, nil
		}
	}

	err = r.Update(context.Background(), instance)
	if err != nil {
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileBucket) bucketNameForInstance(instance *assetstorev1alpha1.Bucket) string {
	return fmt.Sprintf("%s-%s", instance.Namespace, instance.Name)
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
