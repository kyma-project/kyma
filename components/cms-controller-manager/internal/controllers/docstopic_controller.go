package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/internal/handler/docstopic"
	"github.com/kyma-project/kyma/components/cms-controller-manager/internal/webhookconfig"
	cmsv1alpha1 "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DocsTopicReconciler reconciles a DocsTopic object
type DocsTopicReconciler struct {
	client.Client
	Log logr.Logger

	relistInterval   time.Duration
	recorder         record.EventRecorder
	assetSvc         docstopic.AssetService
	bucketSvc        docstopic.BucketService
	webhookConfigSvc webhookconfig.AssetWebhookConfigService
}

type DocsTopicConfig struct {
	RelistInterval time.Duration `envconfig:"default=5m"`
	BucketRegion   string        `envconfig:"-"`
}

func NewDocsTopic(config DocsTopicConfig, log logr.Logger, mgr ctrl.Manager, webhookConfigSvc webhookconfig.AssetWebhookConfigService) *DocsTopicReconciler {
	assetService := newAssetService(mgr.GetClient(), mgr.GetScheme())
	bucketService := newBucketService(mgr.GetClient(), mgr.GetScheme(), config.BucketRegion)

	return &DocsTopicReconciler{
		Client:           mgr.GetClient(),
		Log:              log,
		relistInterval:   config.RelistInterval,
		recorder:         mgr.GetEventRecorderFor("docstopic-controller"),
		assetSvc:         assetService,
		bucketSvc:        bucketService,
		webhookConfigSvc: webhookConfigSvc,
	}
}

// Reconcile reads that state of the cluster for a DocsTopic object and makes changes based on the state read
// Automatically generate RBAC rules to allow the Controller to read and write DocsTopics, Assets, and Buckets
// +kubebuilder:rbac:groups=cms.kyma-project.io,resources=docstopics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cms.kyma-project.io,resources=docstopics/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=assets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=assets/status,verbs=get;list
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=buckets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=assetstore.kyma-project.io,resources=buckets/status,verbs=get;list
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;watch

func (r *DocsTopicReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instance := &cmsv1alpha1.DocsTopic{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	docsTopicLogger := r.Log.WithValues("kind", instance.GetObjectKind().GroupVersionKind().Kind, "namespace", instance.GetNamespace(), "name", instance.GetName())
	commonHandler := docstopic.New(docsTopicLogger, r.recorder, r.assetSvc, r.bucketSvc, r.webhookConfigSvc)
	commonStatus, err := commonHandler.Handle(ctx, instance, instance.Spec.CommonDocsTopicSpec, instance.Status.CommonDocsTopicStatus)
	if updateErr := r.updateStatus(ctx, instance, commonStatus); updateErr != nil {
		finalErr := updateErr
		if err != nil {
			finalErr = errors.Wrapf(err, "along with update error %s", updateErr.Error())
		}
		return ctrl.Result{}, finalErr
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: r.relistInterval,
	}, nil
}

func (r *DocsTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cmsv1alpha1.DocsTopic{}).
		Owns(&v1alpha2.Asset{}).
		Complete(r)
}

func (r *DocsTopicReconciler) updateStatus(ctx context.Context, instance *cmsv1alpha1.DocsTopic, commonStatus *cmsv1alpha1.CommonDocsTopicStatus) error {
	if commonStatus == nil {
		return nil
	}

	copy := instance.DeepCopy()
	copy.Status.CommonDocsTopicStatus = *commonStatus

	return r.Status().Update(ctx, copy)
}
