package serverless

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

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
	Log               *zap.SugaredLogger
	client            resource.Client
	recorder          record.EventRecorder
	config            FunctionConfig
	scheme            *runtime.Scheme
	gitOperator       GitOperator
	statsCollector    StatsCollector
	healthCh          chan bool
	initStateFunction stateFn
}

func NewFunction(client resource.Client, log *zap.SugaredLogger, config FunctionConfig, gitOperator GitOperator, recorder record.EventRecorder, statsCollector StatsCollector, healthCh chan bool) *FunctionReconciler {
	return &FunctionReconciler{
		Log:               log.Named("controllers").Named("function"),
		client:            client,
		recorder:          recorder,
		config:            config,
		gitOperator:       gitOperator,
		healthCh:          healthCh,
		statsCollector:    statsCollector,
		initStateFunction: stateFnInitialize,
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
		WithEventFilter(predicate.Funcs{UpdateFunc: func(event event.UpdateEvent) bool {
			fmt.Println("old:", event.ObjectOld.GetName())
			fmt.Println("new:", event.ObjectNew.GetName())
			oldFn, ok := event.ObjectOld.(*serverlessv1alpha1.Function)
			if !ok {
				fmt.Println("Can't cast ")
				return true
			}
			if oldFn == nil {
				oldFn = &serverlessv1alpha1.Function{}
			}

			newFn, ok := event.ObjectNew.(*serverlessv1alpha1.Function)
			if !ok {
				fmt.Println("Can't cast ")
				return true
			}
			if newFn == nil {
				newFn = &serverlessv1alpha1.Function{}
			}

			equalStasus := equalFunctionStatus(oldFn.Status, newFn.Status)
			fmt.Println("Statuses are equal: ", equalStasus)

			return equalStasus
		}}).
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

	var instance serverlessv1alpha1.Function

	err := r.client.Get(ctx, request.NamespacedName, &instance)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !instance.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	zapLog := r.Log.With(
		"kind", instance.GetObjectKind().GroupVersionKind().Kind,
		"name", instance.GetName(),
		"namespace", instance.GetNamespace(),
		"version", instance.GetGeneration())

	dockerCfg, err := r.readDockerConfig(ctx, &instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	stateReconciler := reconciler{
		fn:  r.initStateFunction,
		log: zapLog,
		k8s: k8s{
			client:         r.client,
			recorder:       r.recorder,
			statsCollector: r.statsCollector,
		},
		cfg: cfg{
			fn:     r.config,
			docker: dockerCfg,
		},
		operator: r.gitOperator,
	}

	stateReconciler.result = ctrl.Result{
		Requeue:      true,
		RequeueAfter: time.Second,
	}

	return stateReconciler.reconcile(ctx, instance)
}

func (r *FunctionReconciler) readDockerConfig(ctx context.Context, instance *serverlessv1alpha1.Function) (DockerConfig, error) {
	var secret corev1.Secret
	// try reading user config
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: r.config.ImageRegistryExternalDockerConfigSecretName}, &secret); err == nil {
		data := readSecretData(secret.Data)
		return DockerConfig{
			ActiveRegistryConfigSecretName: r.config.ImageRegistryExternalDockerConfigSecretName,
			PushAddress:                    data["registryAddress"],
			PullAddress:                    data["registryAddress"],
		}, nil
	}

	// try reading default config
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: r.config.ImageRegistryDefaultDockerConfigSecretName}, &secret); err == nil {
		data := readSecretData(secret.Data)
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
