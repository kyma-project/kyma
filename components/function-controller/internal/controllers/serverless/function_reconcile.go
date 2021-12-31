package serverless

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/kubernetes"
	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
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

func (r *FunctionReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if IsHealthCheckRequest(request) {
		r.healthCh <- true
		return ctrl.Result{}, nil
	}
	instance := &serverlessv1alpha1.Function{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log := r.Log.WithValues("kind", instance.GetObjectKind().GroupVersionKind().Kind,
		"name", instance.GetName(),
		"namespace", instance.GetNamespace(),
		"version", instance.GetGeneration())

	if !instance.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	var configMaps corev1.ConfigMapList
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), r.internalFunctionLabels(instance), &configMaps); err != nil {
		log.Error(err, "Cannot list ConfigMaps")
		return ctrl.Result{}, err
	}

	var jobs batchv1.JobList
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), r.internalFunctionLabels(instance), &jobs); err != nil {
		log.Error(err, "Cannot list Jobs")
		return ctrl.Result{}, err
	}

	var runtimeConfigMap corev1.ConfigMapList
	labels := map[string]string{
		kubernetes.ConfigLabel:  kubernetes.RuntimeLabelValue,
		kubernetes.RuntimeLabel: string(instance.Spec.Runtime),
	}
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), labels, &runtimeConfigMap); err != nil {
		log.Error(err, "Cannot list runtime configmap")
		return ctrl.Result{}, err
	}

	if len(runtimeConfigMap.Items) != 1 {
		return ctrl.Result{}, fmt.Errorf("Expected one config map, found %d, with labels: %+v", len(runtimeConfigMap.Items), labels)
	}

	var deployments appsv1.DeploymentList
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), r.internalFunctionLabels(instance), &deployments); err != nil {
		log.Error(err, "Cannot list Deployments")
		return ctrl.Result{}, err
	}

	var services corev1.ServiceList
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), r.internalFunctionLabels(instance), &services); err != nil {
		log.Error(err, "Cannot list Services")
		return ctrl.Result{}, err
	}

	var hpas autoscalingv1.HorizontalPodAutoscalerList
	if err := r.client.ListByLabel(ctx, instance.GetNamespace(), r.internalFunctionLabels(instance), &hpas); err != nil {
		log.Error(err, "Cannot list HorizontalPodAutoscalers")
		return ctrl.Result{}, err
	}

	dockerConfig, err := r.readDockerConfig(ctx, instance)
	if err != nil {
		log.Error(err, "Cannot read Docker registry configuration")
		return ctrl.Result{}, err
	}

	gitOptions, err := r.readGITOptions(ctx, instance)
	if err != nil {
		return r.updateStatusWithoutRepository(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonSourceUpdateFailed,
			Message:            fmt.Sprintf("Reading git options failed: %v", err),
		})
	}

	revision, err := r.syncRevision(instance, gitOptions)
	if err != nil {
		result, errMsg := NextRequeue(err)
		return r.updateStatusWithoutRepository(ctx, result, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonSourceUpdateFailed,
			Message:            errMsg,
		})
	}

	rtmCfg := fnRuntime.GetRuntimeConfig(instance.Spec.Runtime)
	rtm := fnRuntime.GetRuntime(instance.Spec.Runtime)

	switch {
	case instance.Spec.Type == serverlessv1alpha1.SourceTypeGit && r.isOnSourceChange(instance, revision):
		return r.onSourceChange(ctx, instance, &serverlessv1alpha1.Repository{
			Reference: instance.Spec.Reference,
			BaseDir:   instance.Spec.Repository.BaseDir,
		}, revision)
	case instance.Spec.Type != serverlessv1alpha1.SourceTypeGit && r.isOnConfigMapChange(instance, rtm, configMaps.Items, deployments.Items, dockerConfig):
		return r.onConfigMapChange(ctx, log, instance, rtm, configMaps.Items)
	case instance.Spec.Type == serverlessv1alpha1.SourceTypeGit && r.isOnJobChange(instance, rtmCfg, jobs.Items, deployments.Items, gitOptions, dockerConfig):
		return r.onGitJobChange(ctx, log, instance, rtmCfg, jobs.Items, gitOptions, dockerConfig)
	case instance.Spec.Type != serverlessv1alpha1.SourceTypeGit && r.isOnJobChange(instance, rtmCfg, jobs.Items, deployments.Items, git.Options{}, dockerConfig):
		return r.onJobChange(ctx, log, instance, rtmCfg, configMaps.Items[0].GetName(), jobs.Items, dockerConfig)
	case r.isOnDeploymentChange(instance, rtmCfg, deployments.Items, dockerConfig):
		return r.onDeploymentChange(ctx, log, instance, rtmCfg, deployments.Items, dockerConfig)
	case r.isOnServiceChange(instance, services.Items):
		return r.onServiceChange(ctx, log, instance, services.Items)
	case r.isOnHorizontalPodAutoscalerChange(instance, hpas.Items, deployments.Items):
		return r.onHorizontalPodAutoscalerChange(ctx, log, instance, hpas.Items, deployments.Items[0].GetName())
	default:
		return r.updateDeploymentStatus(ctx, log, instance, deployments.Items, corev1.ConditionTrue)
	}
}

func (r *FunctionReconciler) isOnSourceChange(instance *serverlessv1alpha1.Function, commit string) bool {
	return instance.Status.Commit == "" ||
		commit != instance.Status.Commit ||
		instance.Spec.Reference != instance.Status.Reference ||
		serverlessv1alpha1.RuntimeExtended(instance.Spec.Runtime) != instance.Status.Runtime ||
		instance.Spec.BaseDir != instance.Status.BaseDir ||
		r.getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady) == corev1.ConditionFalse
}

func (r *FunctionReconciler) onSourceChange(ctx context.Context, instance *serverlessv1alpha1.Function, repository *serverlessv1alpha1.Repository, commit string) (ctrl.Result, error) {
	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonSourceUpdated,
		Message:            fmt.Sprintf("Sources %s updated", instance.Name),
	}, repository, commit)
}

func (r *FunctionReconciler) readDockerConfig(ctx context.Context, instance *serverlessv1alpha1.Function) (DockerConfig, error) {
	var secret corev1.Secret
	// try reading user config
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: r.config.ImageRegistryExternalDockerConfigSecretName}, &secret); err == nil {
		data := r.readSecretData(secret.Data)
		return DockerConfig{
			ActiveRegistryConfigSecretName: r.config.ImageRegistryExternalDockerConfigSecretName,
			PushAddress:                    data["registryAddress"],
			PullAddress:                    data["registryAddress"],
		}, nil
	}

	// try reading default config
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: r.config.ImageRegistryDefaultDockerConfigSecretName}, &secret); err == nil {
		data := r.readSecretData(secret.Data)
		if data["isInternal"] == "true" {
			return DockerConfig{
				ActiveRegistryConfigSecretName: r.config.ImageRegistryDefaultDockerConfigSecretName,
				PushAddress:                    data["registryAddress"],
				PullAddress:                    data["serverAddress"],
			}, nil
		} else {
			return DockerConfig{
				ActiveRegistryConfigSecretName: r.config.ImageRegistryDefaultDockerConfigSecretName,
				PushAddress:                    data["registryAddress"],
				PullAddress:                    data["registryAddress"],
			}, nil
		}
	}

	return DockerConfig{}, errors.Errorf("Docker registry configuration not found, none of configuration secrets (%s, %s) found in function namespace", r.config.ImageRegistryDefaultDockerConfigSecretName, r.config.ImageRegistryExternalDockerConfigSecretName)
}

func (r *FunctionReconciler) readSecretData(data map[string][]byte) map[string]string {
	output := make(map[string]string)
	for k, v := range data {
		output[k] = string(v)
	}
	return output
}
