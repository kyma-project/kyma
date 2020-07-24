package serverless

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/kubernetes"
	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

type FunctionReconciler struct {
	Log      logr.Logger
	client   resource.Client
	recorder record.EventRecorder
	config   FunctionConfig
	scheme   *runtime.Scheme
}

func NewFunction(client resource.Client, log logr.Logger, config FunctionConfig, recorder record.EventRecorder) *FunctionReconciler {
	return &FunctionReconciler{
		client:   client,
		Log:      log.WithName("controllers").WithName("function"),
		config:   config,
		recorder: recorder,
	}
}

func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("function-controller").
		For(&serverlessv1alpha1.Function{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&batchv1.Job{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{}).
		Complete(r)
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read and what is in the Function.Spec
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions/status,verbs=get;update;patch
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

	instance := &serverlessv1alpha1.Function{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log := r.Log.WithValues("kind", instance.GetObjectKind().GroupVersionKind().Kind, "name", instance.GetName(), "namespace", instance.GetNamespace(), "version", instance.GetGeneration())

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
		kubernetes.ConfigLabel:  "runtime",
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
		log.Error(err, "Cannot list HorizotalPodAutoscalers")
		return ctrl.Result{}, err
	}

	rtmCfg := fnRuntime.GetRuntimeConfig(instance.Spec.Runtime)
	rtm := fnRuntime.GetRuntime(instance.Spec.Runtime)

	switch {
	case r.isOnConfigMapChange(instance, rtm, configMaps.Items, deployments.Items):
		return r.onConfigMapChange(ctx, log, instance, rtm, configMaps.Items)
	case r.isOnJobChange(ctx, instance, rtmCfg, jobs.Items, deployments.Items):
		return r.onJobChange(ctx, log, instance, rtmCfg, configMaps.Items[0].GetName(), jobs.Items)
	case r.isOnDeploymentChange(ctx, instance, rtmCfg, deployments.Items):
		return r.onDeploymentChange(ctx, log, instance, rtmCfg, deployments.Items)
	case r.isOnServiceChange(instance, services.Items):
		return r.onServiceChange(ctx, log, instance, services.Items)
	case r.isOnHorizontalPodAutoscalerChange(instance, hpas.Items, deployments.Items):
		return r.onHorizontalPodAutoscalerChange(ctx, log, instance, hpas.Items, deployments.Items[0].GetName())
	default:
		return r.updateDeploymentStatus(ctx, log, instance, deployments.Items, corev1.ConditionTrue)
	}
}
