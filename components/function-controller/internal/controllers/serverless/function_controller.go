package serverless

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis"
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

func (r *FunctionReconciler) isOnConfigMapChange(instance *serverlessv1alpha1.Function, configMaps []corev1.ConfigMap, service *servingv1.Service) bool {
	image := r.buildExternalImageAddress(instance)
	if service != nil && service.Spec.Template.Spec.Containers[0].Image == image {
		return false
	}

	return len(configMaps) != 1 ||
		instance.Spec.Source != configMaps[0].Data[configMapFunction] ||
		r.sanitizeDependencies(instance.Spec.Deps) != configMaps[0].Data[configMapDeps] ||
		configMaps[0].Data[configMapHandler] != configMapHandler
}

func (r *FunctionReconciler) isOnJobChange(instance *serverlessv1alpha1.Function, jobs []batchv1.Job, service *servingv1.Service) bool {
	image := r.buildExternalImageAddress(instance)
	if service != nil && service.Spec.Template.Spec.Containers[0].Image == image {
		return false
	}

	expectedJob := r.buildJob(instance, "")
	buildStatus := r.getConditionStatus(instance.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)

	return len(jobs) != 1 ||
		len(jobs[0].Spec.Template.Spec.Containers) != 1 ||
		// Compare image argument
		!r.equalJobs(jobs[0], expectedJob) ||
		buildStatus == corev1.ConditionUnknown ||
		buildStatus == corev1.ConditionFalse
}

func (r *FunctionReconciler) onConfigMapChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, configMaps []corev1.ConfigMap) (ctrl.Result, error) {
	configMapsLen := len(configMaps)

	switch configMapsLen {
	case 0:
		return r.createConfigMap(ctx, log, instance)
	case 1:
		return r.updateConfigMap(ctx, log, instance, configMaps[0])
	default:
		log.Info("To many ConfigMaps for function")
		return r.updateStatus(ctx, ctrl.Result{RequeueAfter: r.config.RequeueDuration}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionConfigurationReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonConfigMapError,
			Message:            "Too many ConfigMaps for function",
		})
	}
}

func (r *FunctionReconciler) createConfigMap(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function) (ctrl.Result, error) {
	configMap := r.buildConfigMap(instance)

	log.Info("Create ConfigMap")
	if err := r.resourceClient.Create(ctx, instance, &configMap); err != nil {
		log.Error(err, "Cannot create ConfigMap")
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("ConfigMap %s created", configMap.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapCreated,
		Message:            fmt.Sprintf("ConfigMap %s created", configMap.GetName()),
	})
}

func (r *FunctionReconciler) updateConfigMap(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, configMap corev1.ConfigMap) (ctrl.Result, error) {
	newConfigMap := configMap.DeepCopy()
	expectedConfigMap := r.buildConfigMap(instance)

	newConfigMap.Data = expectedConfigMap.Data
	newConfigMap.Labels = expectedConfigMap.Labels

	log.Info(fmt.Sprintf("Updating ConfigMap %s", configMap.GetName()))
	if err := r.Client.Update(ctx, newConfigMap); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update ConfigMap %s", newConfigMap.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("ConfigMap %s updated", configMap.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionConfigurationReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonConfigMapUpdated,
		Message:            fmt.Sprintf("ConfigMap %s updated", configMap.GetName()),
	})
}

func (r *FunctionReconciler) onJobChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, configMapName string, jobs []batchv1.Job) (ctrl.Result, error) {
	newJob := r.buildJob(instance, configMapName)
	jobsLen := len(jobs)

	switch {
	case jobsLen == 0:
		return r.createJob(ctx, log, instance, newJob)
	case jobsLen > 1 || !r.equalJobs(jobs[0], newJob):
		return r.deleteJobs(ctx, log, instance)
	default:
		return r.updateBuildStatus(ctx, log, instance, jobs[0])
	}
}

func (r *FunctionReconciler) equalJobs(existing batchv1.Job, expected batchv1.Job) bool {
	// Compare image argument
	return existing.Spec.Template.Spec.Containers[0].Args[0] == expected.Spec.Template.Spec.Containers[0].Args[0]
}

func (r *FunctionReconciler) createJob(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, job batchv1.Job) (ctrl.Result, error) {
	log.Info("Creating Job")
	if err := r.resourceClient.Create(ctx, instance, &job); err != nil {
		log.Error(err, "Cannot create Job")
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Job %s created", job.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonJobCreated,
		Message:            fmt.Sprintf("Job %s created", job.GetName()),
	})
}

func (r *FunctionReconciler) deleteJobs(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function) (ctrl.Result, error) {
	log.Info("Deleting all old jobs")
	selector := apilabels.SelectorFromSet(r.functionLabels(instance))
	if err := r.resourceClient.DeleteAllBySelector(ctx, &batchv1.Job{}, instance.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete old Jobs")
		return ctrl.Result{}, err
	}
	log.Info("Old Jobs deleted")

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionBuildReady,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonJobsDeleted,
		Message:            "Old Jobs deleted",
	})
}

func (r *FunctionReconciler) updateBuildStatus(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, job batchv1.Job) (ctrl.Result, error) {
	switch {
	case job.Status.CompletionTime != nil:
		log.Info(fmt.Sprintf("Job %s finished", job.GetName()))
		return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobFinished,
			Message:            fmt.Sprintf("Job %s finished", job.GetName()),
		})
	case job.Status.Failed < 1:
		log.Info(fmt.Sprintf("Job %s is still in progress", job.GetName()))
		return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobRunning,
			Message:            fmt.Sprintf("Job %s is still in progress", job.GetName()),
		})
	default:
		log.Info(fmt.Sprintf("Job %s failed", job.GetName()))
		return r.updateStatus(ctx, ctrl.Result{RequeueAfter: r.config.RequeueDuration}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionBuildReady,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonJobFailed,
			Message:            fmt.Sprintf("Job %s failed", job.GetName()),
		})
	}
}

func (r *FunctionReconciler) onServiceChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, service *servingv1.Service, revisions []servingv1.Revision) (ctrl.Result, error) {
	newService := r.buildService(log, instance, service)
	revisionLen := len(revisions)

	switch {
	case service == nil:
		return r.createService(ctx, log, instance, newService)
	case !r.equalServices(service, newService):
		return r.updateService(ctx, log, instance, service, newService)
	case revisionLen > 1 && service.Status.IsReady():
		return r.deleteRevisions(ctx, log, instance, service, revisions)
	default:
		return r.updateDeployStatus(ctx, log, instance, service)
	}
}

func (r *FunctionReconciler) createService(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, service servingv1.Service) (ctrl.Result, error) {
	log.Info(fmt.Sprintf("Creating Service %s", service.GetName()))
	if err := r.resourceClient.Create(ctx, instance, &service); err != nil {
		log.Error(err, fmt.Sprintf("Cannot create Service with name %s", service.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Service %s created", service.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonServiceCreated,
		Message:            fmt.Sprintf("Service %s created", service.GetName()),
	})
}

func (r *FunctionReconciler) updateService(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, oldService *servingv1.Service, newService servingv1.Service) (ctrl.Result, error) {
	service := oldService.DeepCopy()
	service.Spec = newService.Spec
	service.ObjectMeta.Labels = newService.GetLabels()

	log.Info(fmt.Sprintf("Updating Service %s", service.GetName()))
	if err := r.Client.Update(ctx, service); err != nil {
		log.Error(err, fmt.Sprintf("Cannot update Service with name %s", service.GetName()))
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Service %s updated", service.GetName()))

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonServiceUpdated,
		Message:            fmt.Sprintf("Service %s updated", service.GetName()),
	})
}

func (r *FunctionReconciler) deleteRevisions(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, service *servingv1.Service, revisions []servingv1.Revision) (ctrl.Result, error) {
	log.Info("Deleting all old revisions")

	selector, err := r.getOldRevisionLabel(instance, revisions)
	if err != nil {
		log.Error(err, "Cannot create proper selector for old revisions")
		return ctrl.Result{}, err
	}

	if err := r.resourceClient.DeleteAllBySelector(ctx, &servingv1.Revision{}, instance.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete old Revisions")
		return ctrl.Result{}, err
	}
	log.Info("Old Revisions deleted")

	return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
		Type:               serverlessv1alpha1.ConditionRunning,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             serverlessv1alpha1.ConditionReasonRevisionsDeleted,
		Message:            "Old Revisions deleted",
	})
}

func (r *FunctionReconciler) updateDeployStatus(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, service *servingv1.Service) (ctrl.Result, error) {
	switch {
	case service.Status.IsReady():
		log.Info(fmt.Sprintf("Service %s is ready", service.GetName()))
		return r.updateStatus(ctx, ctrl.Result{RequeueAfter: r.config.RequeueDuration}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonServiceReady,
			Message:            fmt.Sprintf("Function %s is ready", service.GetName()),
		})

	case r.isServiceInProgress(service):
		log.Info(fmt.Sprintf("Service %s is not ready yet", service.GetName()))
		return r.updateStatus(ctx, ctrl.Result{}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonServiceWaiting,
			Message:            fmt.Sprintf("Service %s is not ready yet", service.GetName()),
		})
	default:
		log.Info(fmt.Sprintf("Service %s failed", service.GetName()))
		return r.updateStatus(ctx, ctrl.Result{RequeueAfter: r.config.RequeueDuration}, instance, serverlessv1alpha1.Condition{
			Type:               serverlessv1alpha1.ConditionRunning,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha1.ConditionReasonServiceFailed,
			Message:            fmt.Sprintf("Service %s failed", service.GetName()),
		})
	}
}

func (r *FunctionReconciler) isServiceInProgress(service *servingv1.Service) bool {
	for _, condition := range service.Status.Conditions {
		if condition.Type == apis.ConditionReady {
			return condition.Status == corev1.ConditionUnknown
		}
	}

	return true
}

func (r *FunctionReconciler) equalServices(existing *servingv1.Service, expected servingv1.Service) bool {
	return existing != nil &&
		len(existing.Spec.Template.Spec.Containers) == len(expected.Spec.Template.Spec.Containers) &&
		existing.Spec.Template.Spec.Containers[0].Image == expected.Spec.Template.Spec.Containers[0].Image &&
		r.envsEqual(existing.Spec.Template.Spec.Containers[0].Env, expected.Spec.Template.Spec.Containers[0].Env) &&
		r.mapsEqual(existing.GetLabels(), expected.GetLabels()) &&
		r.mapsEqual(existing.Spec.Template.GetLabels(), expected.Spec.Template.GetLabels()) &&
		r.mapsEqual(existing.Spec.Template.GetAnnotations(), expected.Spec.Template.GetAnnotations()) &&
		existing.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] == expected.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] &&
		existing.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] == expected.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] &&
		existing.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] == expected.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] &&
		existing.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] == expected.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory]
}

func (r *FunctionReconciler) mapsEqual(existing, expected map[string]string) bool {
	if len(existing) != len(expected) {
		return false
	}

	for key, value := range existing {
		if expected[key] != value {
			return false
		}
	}

	return true
}

func (r *FunctionReconciler) envsEqual(existing, expected []corev1.EnvVar) bool {
	if len(existing) != len(expected) {
		return false
	}
	for key, value := range existing {
		expectedValue := expected[key]

		if expectedValue.Name != value.Name || expectedValue.Value != value.Value || expectedValue.ValueFrom != value.ValueFrom {
			return false
		}
	}

	return true
}

func (r *FunctionReconciler) calculateImageTag(instance *serverlessv1alpha1.Function) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%s", instance.GetUID(), instance.Spec.Source, instance.Spec.Deps)))

	return fmt.Sprintf("%x", hash)
}

func (r *FunctionReconciler) updateStatus(ctx context.Context, result ctrl.Result, instance *serverlessv1alpha1.Function, condition serverlessv1alpha1.Condition) (ctrl.Result, error) {
	condition.LastTransitionTime = metav1.Now()

	service := instance.DeepCopy()
	service.Status.Conditions = r.updateCondition(service.Status.Conditions, condition)

	if r.equalConditions(instance.Status.Conditions, service.Status.Conditions) {
		return result, nil
	}

	if err := r.Status().Update(ctx, service); err != nil {
		return ctrl.Result{}, err
	}

	eventType := "Normal"
	if condition.Status == corev1.ConditionFalse {
		eventType = "Warning"
	}

	r.recorder.Event(instance, eventType, string(condition.Reason), condition.Message)

	return result, nil
}

func (r *FunctionReconciler) updateCondition(conditions []serverlessv1alpha1.Condition, condition serverlessv1alpha1.Condition) []serverlessv1alpha1.Condition {
	conditionTypes := make(map[serverlessv1alpha1.ConditionType]interface{}, 3)
	var result []serverlessv1alpha1.Condition

	result = append(result, condition)
	conditionTypes[condition.Type] = nil

	for _, value := range conditions {
		if _, ok := conditionTypes[value.Type]; ok == false {
			result = append(result, value)
			conditionTypes[value.Type] = nil
		}
	}

	return result
}

func (r *FunctionReconciler) equalConditions(existing, expected []serverlessv1alpha1.Condition) bool {
	if len(existing) != len(expected) {
		return false
	}

	existingMap := make(map[serverlessv1alpha1.ConditionType]serverlessv1alpha1.Condition, len(existing))
	for _, value := range existing {
		existingMap[value.Type] = value
	}

	for _, value := range expected {
		if existingMap[value.Type].Status != value.Status || existingMap[value.Type].Reason != value.Reason || existingMap[value.Type].Message != value.Message {
			return false
		}
	}

	return true
}

func (r *FunctionReconciler) getConditionStatus(conditions []serverlessv1alpha1.Condition, conditionType serverlessv1alpha1.ConditionType) corev1.ConditionStatus {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status
		}
	}

	return corev1.ConditionUnknown
}

func (r *FunctionReconciler) functionLabels(instance *serverlessv1alpha1.Function) map[string]string {
	labels := make(map[string]string, len(instance.GetLabels())+3)
	for key, value := range instance.GetLabels() {
		labels[key] = value
	}

	labels[functionNameLabel] = instance.Name
	labels[functionManagedByLabel] = "function-controller"
	labels[functionUUIDLabel] = string(instance.GetUID())

	return labels
}

func (r *FunctionReconciler) servingPodLabels(log logr.Logger, instance *serverlessv1alpha1.Function, bindingAnnotation string) map[string]string {
	functionLabels := r.functionLabels(instance)
	if bindingAnnotation == "" {
		return functionLabels
	}

	type binding map[string]map[string]map[string]string
	var bindings binding
	if err := json.Unmarshal([]byte(bindingAnnotation), &bindings); err != nil {
		log.Error(err, fmt.Sprintf("Cannot parse SeriveBindingUsage annotation %s", bindingAnnotation))
	}

	for _, service := range bindings {
		for key, value := range service["injectedLabels"] {
			functionLabels[key] = value
		}
	}

	return functionLabels
}

func (r *FunctionReconciler) buildInternalImageAddress(instance *serverlessv1alpha1.Function) string {
	imageTag := r.calculateImageTag(instance)
	return fmt.Sprintf("%s/%s-%s:%s", r.config.Docker.Address, instance.Namespace, instance.Name, imageTag)
}

func (r *FunctionReconciler) buildExternalImageAddress(instance *serverlessv1alpha1.Function) string {
	imageTag := r.calculateImageTag(instance)
	return fmt.Sprintf("%s/%s-%s:%s", r.config.Docker.ExternalAddress, instance.Namespace, instance.Name, imageTag)
}

func (r *FunctionReconciler) sanitizeDependencies(dependencies string) string {
	result := "{}"
	if strings.Trim(dependencies, " ") != "" {
		result = dependencies
	}

	return result
}

func (r *FunctionReconciler) buildJob(instance *serverlessv1alpha1.Function, configMapName string) batchv1.Job {
	imageName := r.buildInternalImageAddress(instance)
	one := int32(1)
	zero := int32(0)

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-build-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
			Labels:       r.functionLabels(instance),
		},
		Spec: batchv1.JobSpec{
			Parallelism:           &one,
			Completions:           &one,
			ActiveDeadlineSeconds: nil,
			BackoffLimit:          &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: r.functionLabels(instance),
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "sources",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: configMapName},
								},
							},
						},
						{
							Name: "runtime",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: r.config.Build.RuntimeConfigMapName},
								},
							},
						},
						{
							Name: "credentials",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{SecretName: r.config.ImagePullSecretName},
							},
						},
						{
							Name:         "tekton-home",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
						},
						{
							Name:         "tekton-workspace",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:    "credential-initializer",
							Image:   r.config.Build.CredsInitImage,
							Command: []string{"/ko-app/creds-init"},
							Args:    []string{fmt.Sprintf("-basic-docker=credentials=http://%s", imageName)},
							Env: []corev1.EnvVar{
								{Name: "HOME", Value: "/tekton/home"},
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "tekton-home", ReadOnly: false, MountPath: "/tekton/home"},
								{Name: "credentials", ReadOnly: false, MountPath: "/tekton/creds-secrets/credentials"},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "executor",
							Image: r.config.Build.ExecutorImage,
							Args:  []string{fmt.Sprintf("--destination=%s", imageName), "--insecure", "--skip-tls-verify"},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: r.config.Build.LimitsMemoryValue,
									corev1.ResourceCPU:    r.config.Build.LimitsCPUValue,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: r.config.Build.RequestsMemoryValue,
									corev1.ResourceCPU:    r.config.Build.RequestsCPUValue,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "sources", ReadOnly: true, MountPath: "/src"},
								{Name: "runtime", ReadOnly: true, MountPath: "/workspace"},
								{Name: "tekton-home", ReadOnly: false, MountPath: "/tekton/home"},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "DOCKER_CONFIG", Value: "/tekton/home/.docker/"},
							},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: r.config.ImagePullAccountName,
				},
			},
		},
	}

	return job
}

func (r *FunctionReconciler) buildService(log logr.Logger, instance *serverlessv1alpha1.Function, oldService *servingv1.Service) servingv1.Service {
	imageName := r.buildExternalImageAddress(instance)
	annotations := map[string]string{
		"autoscaling.knative.dev/minScale": "1",
		"autoscaling.knative.dev/maxScale": "1",
	}
	if instance.Spec.MinReplicas != nil {
		annotations["autoscaling.knative.dev/minScale"] = fmt.Sprintf("%d", *instance.Spec.MinReplicas)
	}
	if instance.Spec.MaxReplicas != nil {
		annotations["autoscaling.knative.dev/maxScale"] = fmt.Sprintf("%d", *instance.Spec.MaxReplicas)
	}
	serviceLabels := r.functionLabels(instance)
	serviceLabels["serving.knative.dev/visibility"] = "cluster-local"

	bindingAnnotation := ""
	if oldService != nil {
		bindingAnnotation = oldService.GetAnnotations()[serviceBindingUsagesAnnotation]
	}
	podLabels := r.servingPodLabels(log, instance, bindingAnnotation)

	service := servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName(),
			Namespace: instance.GetNamespace(),
			Labels:    serviceLabels,
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: annotations,
						Labels:      podLabels,
					},
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "lambda",
									Image:           imageName,
									Env:             append(instance.Spec.Env, envVarsForRevision...),
									Resources:       instance.Spec.Resources,
									ImagePullPolicy: corev1.PullIfNotPresent,
								},
							},
							ServiceAccountName: r.config.ImagePullAccountName,
						},
					},
				},
			},
		},
	}

	return service
}

func getMaxGeneration(revisions []servingv1.Revision) (int, error) {
	maxGeneration := -1
	for _, revision := range revisions {
		generationString, ok := revision.Labels[cfgGenerationLabel]
		if !ok {
			// todo extract to var
			return -1, errors.New(fmt.Sprintf("Revision %s in namespace %s doesn't have %s label", revision.Name, revision.Namespace, cfgGenerationLabel))
		}
		generation, err := strconv.Atoi(generationString)
		if err != nil {
			// todo extract to var
			return -1, errors.New(fmt.Sprintf("Couldn't convert label key %s to number, revision %s in namespace %s", generationString, revision.Name, revision.Namespace))
		}
		if generation > maxGeneration {
			maxGeneration = generation
		}
	}
	return maxGeneration, nil
}

func (r *FunctionReconciler) getOldRevisionLabel(instance *serverlessv1alpha1.Function, revisions []servingv1.Revision) (apilabels.Selector, error) {
	maxGen, err := getMaxGeneration(revisions)
	if err != nil {
		return nil, err
	}

	selector := apilabels.NewSelector()
	uuidReq, err := apilabels.NewRequirement(functionUUIDLabel, selection.Equals, []string{string(instance.UID)})
	if err != nil {
		return nil, err
	}
	generationReq, err := apilabels.NewRequirement(cfgGenerationLabel, selection.NotEquals, []string{strconv.Itoa(maxGen)})
	if err != nil {
		return nil, err
	}

	return selector.Add(*uuidReq, *generationReq), nil
}

func (r *FunctionReconciler) buildConfigMap(instance *serverlessv1alpha1.Function) corev1.ConfigMap {
	data := map[string]string{
		configMapHandler:  configMapHandler,
		configMapFunction: instance.Spec.Source,
		configMapDeps:     r.sanitizeDependencies(instance.Spec.Deps),
	}
	labels := r.functionLabels(instance)

	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       labels,
			GenerateName: fmt.Sprintf("%s-", instance.GetName()),
			Namespace:    instance.GetNamespace(),
		},
		Data: data,
	}
}
