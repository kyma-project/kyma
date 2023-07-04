package serverless

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

const (
	healthCheckTimeout  = time.Second
	keyRegistryPullAddr = "pullRegAddr"
	keyRegistryPushAddr = "pushRegAddr"
	keyRegistryAddress  = "registryAddress"
	keyIsInternal       = "isInternal"
)

//go:generate mockery --name=GitClient --output=automock --outpkg=automock --case=underscore
type GitClient interface {
	LastCommit(options git.Options) (string, error)
	Clone(path string, options git.Options) (string, error)
}

//go:generate mockery --name=GitClientFactory --output=automock --outpkg=automock --case=underscore
type GitClientFactory interface {
	GetGitClient(logger *zap.SugaredLogger) git.GitClient
}

//go:generate mockery --name=StatsCollector --output=automock --outpkg=automock --case=underscore
type StatsCollector interface {
	UpdateReconcileStats(f *serverlessv1alpha2.Function, cond serverlessv1alpha2.Condition)
}

type FunctionReconciler struct {
	Log               *zap.SugaredLogger
	client            resource.Client
	recorder          record.EventRecorder
	config            FunctionConfig
	gitFactory        GitClientFactory
	statsCollector    StatsCollector
	healthCh          chan bool
	initStateFunction stateFn
}

func NewFunctionReconciler(client resource.Client, log *zap.SugaredLogger, config FunctionConfig, gitFactory GitClientFactory, recorder record.EventRecorder, statsCollector StatsCollector, healthCh chan bool) *FunctionReconciler {
	return &FunctionReconciler{
		Log:               log,
		client:            client,
		recorder:          recorder,
		config:            config,
		gitFactory:        gitFactory,
		healthCh:          healthCh,
		statsCollector:    statsCollector,
		initStateFunction: stateFnInitialize,
	}
}

func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		Named("function-controller").
		For(&serverlessv1alpha2.Function{}, builder.WithPredicates(predicate.Funcs{UpdateFunc: IsNotFunctionStatusUpdate(r.Log)})).
		Owns(&corev1.ConfigMap{}).
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
		r.sendHealthCheck()
		return ctrl.Result{}, nil
	}

	r.Log.With(
		"name", request.Name,
		"namespace", request.Namespace).
		Debug("starting pre-reconciliation steps")

	var instance serverlessv1alpha2.Function

	err := r.client.Get(ctx, request.NamespacedName, &instance)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !instance.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	contextLogger := r.Log.With(
		"kind", instance.GetObjectKind().GroupVersionKind().Kind,
		"name", instance.GetName(),
		"namespace", instance.GetNamespace(),
		"version", instance.GetGeneration())

	dockerCfg, err := r.readDockerConfig(ctx, &instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	contextLogger.Debug("starting state machine")

	stateReconciler := reconciler{
		fn:  r.initStateFunction,
		log: contextLogger,
		k8s: k8s{
			client:         r.client,
			recorder:       r.recorder,
			statsCollector: r.statsCollector,
		},
		cfg: cfg{
			fn:     r.config,
			docker: dockerCfg,
		},
		gitClient: r.gitFactory.GetGitClient(contextLogger),
	}

	stateReconciler.result = ctrl.Result{
		RequeueAfter: time.Second * 1,
	}

	return stateReconciler.reconcile(ctx, instance)
}

func (r *FunctionReconciler) sendHealthCheck() {
	r.Log.Debug("health check request received")

	select {
	case r.healthCh <- true:
		r.Log.Debug("health check request responded")
	case <-time.After(healthCheckTimeout):
		r.Log.Warn(errors.New("timeout when responding to health check"))
	}
}

func (r *FunctionReconciler) readDockerConfig(ctx context.Context, instance *serverlessv1alpha2.Function) (DockerConfig, error) {
	var secret corev1.Secret
	// try reading user config
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: r.config.ImageRegistryExternalDockerConfigSecretName}, &secret); err == nil {
		data := readSecretData(secret.Data)
		return DockerConfig{
			ActiveRegistryConfigSecretName: r.config.ImageRegistryExternalDockerConfigSecretName,
			PushAddress:                    data[keyRegistryAddress],
			PullAddress:                    data[keyRegistryAddress],
		}, nil
	}

	// try reading default config
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: r.config.ImageRegistryDefaultDockerConfigSecretName}, &secret); err != nil {
		return DockerConfig{}, errors.Wrapf(err, "docker registry configuration not found, none of configuration secrets (%s, %s) found in function namespace", r.config.ImageRegistryDefaultDockerConfigSecretName, r.config.ImageRegistryExternalDockerConfigSecretName)
	}
	data := readSecretData(secret.Data)
	if data[keyIsInternal] == "true" {
		return DockerConfig{
			ActiveRegistryConfigSecretName: r.config.ImageRegistryDefaultDockerConfigSecretName,
			PushAddress:                    data[keyRegistryPushAddr],
			PullAddress:                    data[keyRegistryPullAddr],
		}, nil
	} else {
		return DockerConfig{
			ActiveRegistryConfigSecretName: r.config.ImageRegistryDefaultDockerConfigSecretName,
			PushAddress:                    data[keyRegistryAddress],
			PullAddress:                    data[keyRegistryAddress],
		}, nil
	}

}
