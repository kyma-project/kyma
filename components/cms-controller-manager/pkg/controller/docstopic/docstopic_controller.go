package docstopic

import (
	"context"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/handler/docstopic"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/record"
	"time"

	assetstoreapi "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	cmsv1alpha1 "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
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

var log = logf.Log.WithName("docstopic-controller")

// Add creates a new DocsTopic Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	cfg, err := loadConfig("APP")
	if err != nil {
		return errors.Wrapf(err, "while loading configuration")
	}

	client := mgr.GetClient()
	scheme := mgr.GetScheme()
	assetService := newAssetService(client, scheme)
	bucketService := newBucketService(client, scheme, cfg.BucketRegion)

	reconciler := &ReconcileDocsTopic{
		relistInterval: cfg.DocsTopicRelistInterval,
		Client:         mgr.GetClient(),
		scheme:         mgr.GetScheme(),
		recorder:       mgr.GetRecorder("docstopic-controller"),
		assetSvc:       assetService,
		bucketSvc:      bucketService,
	}

	return add(mgr, reconciler)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("docstopic-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to DocsTopic
	err = c.Watch(&source.Kind{Type: &cmsv1alpha1.DocsTopic{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &assetstoreapi.Asset{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cmsv1alpha1.DocsTopic{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDocsTopic{}

// ReconcileDocsTopic reconciles a DocsTopic object
type ReconcileDocsTopic struct {
	client.Client
	scheme         *runtime.Scheme
	relistInterval time.Duration
	recorder       record.EventRecorder
	assetSvc       docstopic.AssetService
	bucketSvc      docstopic.BucketService
}

// Reconcile reads that state of the cluster for a DocsTopic object and makes changes based on the state read
// Automatically generate RBAC rules to allow the Controller to read and write DocsTopics, Assets, and Buckets
// +kubebuilder:rbac:groups=cms.kyma-project.io,resources=docstopics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cms.kyma-project.io,resources=docstopics/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=assets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=assets/status,verbs=get;list;update;patch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=buckets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=buckets/status,verbs=get;list;update;patch
func (r *ReconcileDocsTopic) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	instance := &cmsv1alpha1.DocsTopic{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	docsTopicLogger := log.WithValues("kind", instance.GetObjectKind().GroupVersionKind().Kind, "namespace", instance.GetNamespace(), "name", instance.GetName())
	commonHandler := docstopic.New(docsTopicLogger, r.recorder, r.assetSvc, r.bucketSvc)
	commonStatus, err := commonHandler.Handle(ctx, instance, instance.Spec.CommonDocsTopicSpec, instance.Status.CommonDocsTopicStatus)
	if updateErr := r.updateStatus(ctx, instance, commonStatus); updateErr != nil {
		finalErr := updateErr
		if err != nil {
			finalErr = errors.Wrapf(err, "along with update error %s", updateErr.Error())
		}
		return reconcile.Result{}, finalErr
	}
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{
		RequeueAfter: r.relistInterval,
	}, nil
}

func (r *ReconcileDocsTopic) updateStatus(ctx context.Context, instance *cmsv1alpha1.DocsTopic, commonStatus *cmsv1alpha1.CommonDocsTopicStatus) error {
	if commonStatus == nil {
		return nil
	}

	copy := instance.DeepCopy()
	copy.Status.CommonDocsTopicStatus = *commonStatus

	return r.Status().Update(ctx, copy)
}
