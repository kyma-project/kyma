package serverless

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"knative.dev/pkg/apis"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	serviceBindingUsagesAnnotation = "servicebindingusages.servicecatalog.kyma-project.io/tracing-information"

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

func (r *FunctionReconciler) onServiceChange(ctx context.Context, log logr.Logger, instance *serverlessv1alpha1.Function, service *servingv1.Service) (ctrl.Result, error) {
	newService := r.buildService(log, instance, service)

	switch {
	case service == nil:
		return r.createService(ctx, log, instance, newService)
	case !r.equalServices(service, newService):
		return r.updateService(ctx, log, instance, service, newService)
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

	labels[serverlessv1alpha1.FunctionNameLabel] = instance.Name
	labels[serverlessv1alpha1.FunctionManagedByLabel] = "function-controller"
	labels[serverlessv1alpha1.FunctionUUIDLabel] = string(instance.GetUID())

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
