package serverless

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
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
	Clone(path string, options git.Options) (string, error)
}

type FunctionReconciler struct {
	Log         logr.Logger
	client      resource.Client
	recorder    record.EventRecorder
	config      FunctionConfig
	scheme      *runtime.Scheme
	gitOperator GitOperator
}

func NewFunction(client resource.Client, log logr.Logger, config FunctionConfig, recorder record.EventRecorder) *FunctionReconciler {
	return &FunctionReconciler{
		client:      client,
		Log:         log.WithName("controllers").WithName("function"),
		config:      config,
		recorder:    recorder,
		gitOperator: git.New(),
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
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.config.MaxConcurrentReconciles,
		}).
		Complete(r)
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

	activeJobs, err := r.getActiveJobs(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if activeJobs >= r.config.Build.MaxActiveJobs {
		return ctrl.Result{
			RequeueAfter: time.Second * 5,
		}, nil
	}

	instance := &serverlessv1alpha1.Function{}
	err = r.client.Get(ctx, request.NamespacedName, instance)
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
		return r.updateStatusWithoutRepository(ctx, ctrl.Result{
			RequeueAfter: r.config.GitFetchRequeueDuration,
		}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonSourceUpdateFailed,
			Message:            fmt.Sprintf("Sources update failed: %v", err),
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
	case instance.Spec.Type != serverlessv1alpha1.SourceTypeGit && r.isOnConfigMapChange(instance, rtm, configMaps.Items, deployments.Items):
		return r.onConfigMapChange(ctx, log, instance, rtm, configMaps.Items)
	case instance.Spec.Type == serverlessv1alpha1.SourceTypeGit && r.isOnJobChange(instance, rtmCfg, jobs.Items, deployments.Items, gitOptions):
		return r.onGitJobChange(ctx, log, instance, rtmCfg, jobs.Items, gitOptions)
	case instance.Spec.Type != serverlessv1alpha1.SourceTypeGit && r.isOnJobChange(instance, rtmCfg, jobs.Items, deployments.Items, git.Options{}):
		return r.onJobChange(ctx, log, instance, rtmCfg, configMaps.Items[0].GetName(), jobs.Items)
	case r.isOnDeploymentChange(instance, rtmCfg, deployments.Items):
		return r.onDeploymentChange(ctx, log, instance, rtmCfg, deployments.Items)
	case r.isOnServiceChange(instance, services.Items):
		return r.onServiceChange(ctx, log, instance, services.Items)
	case r.isOnHorizontalPodAutoscalerChange(instance, hpas.Items, deployments.Items):
		return r.onHorizontalPodAutoscalerChange(ctx, log, instance, hpas.Items, deployments.Items[0].GetName())
	default:
		return r.updateDeploymentStatus(ctx, log, instance, deployments.Items, corev1.ConditionTrue)
	}
}

func (r *FunctionReconciler) getActiveJobs(ctx context.Context) (int, error) {
	var allJobs batchv1.JobList
	if err := r.client.ListByLabel(ctx, "", map[string]string{serverlessv1alpha1.FunctionManagedByLabel: "function-controller"}, &allJobs); err != nil {
		r.Log.Error(err, "Cannot list Jobs")
		return 0, err
	}

	activeJobs := 0
	for _, j := range allJobs.Items {
		if j.Status.Active > 0 {
			activeJobs++
		}
		r.Log.WithValues("name:", j.Name, "active", j.Status.Active)
	}
	return activeJobs, nil
}

func (r *FunctionReconciler) isOnSourceChange(instance *serverlessv1alpha1.Function, commit string) bool {
	return instance.Status.Commit == "" ||
		commit != instance.Status.Commit ||
		instance.Spec.Reference != instance.Status.Reference ||
		instance.Spec.Runtime != instance.Status.Runtime ||
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

func (r *FunctionReconciler) syncRevision(instance *serverlessv1alpha1.Function, options git.Options) (string, error) {
	if instance.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		return r.gitOperator.LastCommit(options)
	}
	return "", nil
}

func (r *FunctionReconciler) readGITOptions(ctx context.Context, instance *serverlessv1alpha1.Function) (git.Options, error) {
	if instance.Spec.Type != serverlessv1alpha1.SourceTypeGit {
		return git.Options{}, nil
	}

	var gitRepository serverlessv1alpha1.GitRepository
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: instance.Spec.Source}, &gitRepository); err != nil {
		return git.Options{}, err
	}

	var auth *git.AuthOptions
	if gitRepository.Spec.Auth != nil {
		var secret corev1.Secret
		if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: gitRepository.Spec.Auth.SecretName}, &secret); err != nil {
			return git.Options{}, err
		}
		auth = &git.AuthOptions{
			Type:        git.RepositoryAuthType(gitRepository.Spec.Auth.Type),
			Credentials: r.readSecretData(secret.Data),
			SecretName:  gitRepository.Spec.Auth.SecretName,
		}
	}

	if instance.Spec.Reference == "" {
		return git.Options{}, fmt.Errorf("reference has to specified")
	}

	return git.Options{
		URL:       gitRepository.Spec.URL,
		Reference: instance.Spec.Reference,
		Auth:      auth,
	}, nil
}

func (r *FunctionReconciler) readSecretData(data map[string][]byte) map[string]string {
	output := make(map[string]string)
	for k, v := range data {
		output[k] = string(v)
	}
	return output
}
