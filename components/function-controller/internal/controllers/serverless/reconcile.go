package serverless

import (
	"context"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const (
	functionNameLabel      = "serverless.kyma-project.io/function-name"
	functionManagedByLabel = "serverless.kyma-project.io/managed-by"
	functionUUIDLabel      = "serverless.kyma-project.io/uuid"

	serviceBindingUsagesAnnotation = "servicebindingusages.servicecatalog.kyma-project.io/tracing-information"

	cfgGenerationLabel = "serving.knative.dev/configurationGeneration"

	configMapFunction = "handler.js"
	configMapHandler  = "handler.main"
	configMapDeps     = "package.json"
)

var (
	envVarsForRevision = []corev1.EnvVar{
		{Name: "FUNC_HANDLER", Value: "main"},
		{Name: "MOD_NAME", Value: "handler"},
		{Name: "FUNC_TIMEOUT", Value: "180"},
		{Name: "FUNC_RUNTIME", Value: "nodejs12"},
		// {Name: "FUNC_MEMORY_LIMIT", Value: "128Mi"},
		{Name: "FUNC_PORT", Value: "8080"},
		{Name: "NODE_PATH", Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules"},
	}
)

type FunctionReconciler struct {
	client.Client
	Log logr.Logger

	resourceClient resource.Resource
	recorder       record.EventRecorder
	config         FunctionConfig
	scheme         *runtime.Scheme
}

func NewFunction(client client.Client, log logr.Logger, config FunctionConfig, scheme *runtime.Scheme, recorder record.EventRecorder) *FunctionReconciler {
	resourceClient := resource.New(client, scheme)

	return &FunctionReconciler{
		Client:         client,
		Log:            log,
		config:         config,
		resourceClient: resourceClient,
		scheme:         scheme,
		recorder:       recorder,
	}
}

func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessv1alpha1.Function{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&batchv1.Job{}).
		Owns(&servingv1.Service{}).
		Owns(&servingv1.Revision{}).
		Complete(r)
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read and what is in the Function.Spec
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services/status,verbs=get
// +kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups="batch",resources=jobs/status,verbs=get
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *FunctionReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instance := &serverlessv1alpha1.Function{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log := r.Log.WithValues("kind", instance.GetObjectKind().GroupVersionKind().Kind, "name", instance.GetName(), "namespace", instance.GetNamespace(), "version", instance.GetGeneration())

	log.Info("Listing ConfigMaps")
	var configMaps corev1.ConfigMapList
	if err := r.resourceClient.ListByLabel(ctx, instance.GetNamespace(), r.functionLabels(instance), &configMaps); err != nil {
		log.Error(err, "Cannot list ConfigMaps")
		return ctrl.Result{}, err
	}
	log.Info("Listing Jobs")
	var jobs batchv1.JobList
	if err := r.resourceClient.ListByLabel(ctx, instance.GetNamespace(), r.functionLabels(instance), &jobs); err != nil {
		log.Error(err, "Cannot list Jobs")
		return ctrl.Result{}, err
	}
	log.Info("Gathering Service")
	service := &servingv1.Service{}
	if err := r.Client.Get(ctx, request.NamespacedName, service); err != nil {
		if apierrors.IsNotFound(err) {
			service = nil
		} else {
			log.Error(err, "Cannot get Service %s", instance.GetName())
			return ctrl.Result{}, err
		}
	}

	log.Info("Listing Revisions")
	var revisions servingv1.RevisionList
	if err := r.resourceClient.ListByLabel(ctx, instance.GetNamespace(), r.functionLabels(instance), &revisions); err != nil {
		log.Error(err, "Cannot list Revisions")
		return ctrl.Result{}, err
	}

	switch {
	case r.isOnConfigMapChange(instance, configMaps.Items, service):
		return r.onConfigMapChange(ctx, log, instance, configMaps.Items)
	case r.isOnJobChange(instance, jobs.Items, service):
		return r.onJobChange(ctx, log, instance, configMaps.Items[0].GetName(), jobs.Items)
	default:
		return r.onServiceChange(ctx, log, instance, service, revisions.Items)
	}
}
