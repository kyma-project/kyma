package serverless

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/kubernetes"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

//go:generate mockery -name=GitOperator -output=automock -outpkg=automock -case=underscore
type GitOperator interface {
	LastCommit(options git.Options) (string, error)
}

//go:generate mockery -name=StatsCollector -output=automock -outpkg=automock -case=underscore
type StatsCollector interface {
	UpdateReconcileStats(f *serverlessv1alpha1.Function, cond serverlessv1alpha1.Condition)
}

type FunctionReconciler struct {
	Log            logr.Logger
	client         resource.Client
	recorder       record.EventRecorder
	config         FunctionConfig
	scheme         *runtime.Scheme
	gitOperator    GitOperator
	statsCollector StatsCollector
	healthCh       chan bool
}

type functionResources struct {
	configMaps       corev1.ConfigMapList
	jobs             batchv1.JobList
	runtimeConfigMap corev1.ConfigMapList
	deployments      appsv1.DeploymentList
	services         corev1.ServiceList
	hpas             autoscalingv1.HorizontalPodAutoscalerList
}

type statusUpdater struct {
	client         resource.Client
	config         FunctionConfig
	statsCollector StatsCollector
	recorder       record.EventRecorder
}

func NewFunction(client resource.Client, log logr.Logger, config FunctionConfig, gitOperator GitOperator, recorder record.EventRecorder, statsCollector StatsCollector, healthCh chan bool) *FunctionReconciler {
	return &FunctionReconciler{
		Log:            log.WithName("controllers").WithName("function"),
		client:         client,
		recorder:       recorder,
		config:         config,
		gitOperator:    gitOperator,
		healthCh:       healthCh,
		statsCollector: statsCollector,
	}
}

func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		Named("function-controller").
		For(&serverlessv1alpha1.Function{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&batchv1.Job{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{}).
		WithOptions(controller.Options{
			RateLimiter: workqueue.NewMaxOfRateLimiter(
				workqueue.NewItemExponentialFailureRateLimiter(r.config.GitFetchRequeueDuration, 300*time.Second),
				// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
				&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
			),
			MaxConcurrentReconciles: 1, // Build job scheduling mechanism requires this parameter to be set to 1. The mechanism is based on getting active and stateless jobs, concurrent reconciles makes it non deterministic . Value 1 removes data races while fetching list of jobs. https://github.com/kyma-project/kyma/issues/10037
		}).
		Build(r)
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read and what is in the Function.Spec
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=gitrepositories,verbs=get
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups="apps",resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups="batch",resources=jobs/status,verbs=get
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;deletecollection
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;deletecollection
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *FunctionReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	if IsHealthCheckRequest(request) {
		r.healthCh <- true
		return ctrl.Result{}, nil
	}
	instance := &serverlessv1alpha1.Function{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		return ctrl.Result{}, ctrlclient.IgnoreNotFound(err)
	}

	log := r.Log.WithValues("kind", instance.GetObjectKind().GroupVersionKind().Kind,
		"name", instance.GetName(),
		"namespace", instance.GetNamespace(),
		"version", instance.GetGeneration())

	if !instance.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	resources, err := r.fetchFunctionResources(ctx, instance, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	su := r.newStatusUpdater()
	if instance.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		return newGitFunctionReconciler(r).reconcileGitFunction(ctx, instance, resources, su, log)
	}
	return r.reconcileInlineFunctionReconcile(ctx, instance, resources, su, log)
}

func (r *FunctionReconciler) fetchFunctionResources(ctx context.Context, instance *serverlessv1alpha1.Function, log logr.Logger) (*functionResources, error) {
	var configMaps corev1.ConfigMapList
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), internalFunctionLabels(instance), &configMaps); err != nil {
		log.Error(err, "Cannot list ConfigMaps")
		return nil, err
	}

	var jobs batchv1.JobList
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), internalFunctionLabels(instance), &jobs); err != nil {
		log.Error(err, "Cannot list Jobs")
		return nil, err
	}

	var runtimeConfigMap corev1.ConfigMapList
	labels := map[string]string{
		kubernetes.ConfigLabel:  kubernetes.RuntimeLabelValue,
		kubernetes.RuntimeLabel: string(instance.Spec.Runtime),
	}
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), labels, &runtimeConfigMap); err != nil {
		log.Error(err, "Cannot list runtime configmap")
		return nil, err
	}

	if len(runtimeConfigMap.Items) != 1 {
		return nil, fmt.Errorf("Expected one config map, found %d, with labels: %+v", len(runtimeConfigMap.Items), labels)
	}

	var deployments appsv1.DeploymentList
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), internalFunctionLabels(instance), &deployments); err != nil {
		log.Error(err, "Cannot list Deployments")
		return nil, err
	}

	var services corev1.ServiceList
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), internalFunctionLabels(instance), &services); err != nil {
		log.Error(err, "Cannot list Services")
		return nil, err
	}

	var hpas autoscalingv1.HorizontalPodAutoscalerList
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), internalFunctionLabels(instance), &hpas); err != nil {
		log.Error(err, "Cannot list HorizontalPodAutoscalers")
		return nil, err
	}
	return &functionResources{
		configMaps:       configMaps,
		jobs:             jobs,
		runtimeConfigMap: runtimeConfigMap,
		deployments:      deployments,
		services:         services,
		hpas:             hpas,
	}, nil
}

func (r *FunctionReconciler) newStatusUpdater() *statusUpdater {
	return &statusUpdater{
		client:         r.client,
		config:         r.config,
		statsCollector: r.statsCollector,
		recorder:       r.recorder,
	}
}
