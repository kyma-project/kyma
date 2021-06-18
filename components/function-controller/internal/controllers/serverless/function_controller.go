package serverless

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const (
	FunctionSourceKey = "source"
	FunctionDepsKey   = "dependencies"
)

var (
	// shared between all runtimes
	envVarsForDeployment = []corev1.EnvVar{
		{Name: "FUNC_HANDLER", Value: "main"},
		{Name: "MOD_NAME", Value: "handler"},
		{Name: "FUNC_PORT", Value: "8080"},
	}
)

func (r *FunctionReconciler) mapsEqual(existing, expected map[string]string) bool {
	if len(existing) != len(expected) {
		return false
	}

	for key, value := range existing {
		if v, ok := expected[key]; !ok || v != value {
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

		if expectedValue.Name != value.Name || expectedValue.Value != value.Value || expectedValue.ValueFrom.String() != value.ValueFrom.String() { // valueFrom check is by string representation
			return false
		}
	}

	return true
}

func (r *FunctionReconciler) calculateImageTag(instance *serverlessv1alpha1.Function) string {
	hash := sha256.Sum256([]byte(strings.Join([]string{
		string(instance.GetUID()),
		instance.Spec.Source,
		instance.Spec.Deps,
		string(instance.Status.Runtime),
	}, "-")))

	return fmt.Sprintf("%x", hash)
}

func (r *FunctionReconciler) updateStatusWithoutRepository(ctx context.Context, result ctrl.Result, instance *serverlessv1alpha1.Function, condition serverlessv1alpha1.Condition) (ctrl.Result, error) {
	return r.updateStatus(ctx, result, instance, condition, nil, "")
}

func (r *FunctionReconciler) calculateGitImageTag(instance *serverlessv1alpha1.Function) string {
	data := strings.Join([]string{
		string(instance.GetUID()),
		instance.Status.Commit,
		instance.Status.Repository.BaseDir,
		string(instance.Status.Runtime),
	}, "-")
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func (r *FunctionReconciler) updateStatus(ctx context.Context, result ctrl.Result, instance *serverlessv1alpha1.Function, condition serverlessv1alpha1.Condition, repository *serverlessv1alpha1.Repository, commit string) (ctrl.Result, error) {
	condition.LastTransitionTime = metav1.Now()

	service := instance.DeepCopy()
	service.Status.Conditions = r.updateCondition(service.Status.Conditions, condition)

	equalConditions := r.equalConditions(instance.Status.Conditions, service.Status.Conditions)
	if equalConditions && instance.Spec.Type != serverlessv1alpha1.SourceTypeGit {
		return result, nil
	}
	// checking if status changed in gitops flow
	if equalConditions && r.equalRepositories(instance.Status.Repository, repository) &&
		instance.Status.Commit == commit {
		return result, nil
	}

	if repository != nil {
		service.Status.Repository = *repository
		service.Status.Commit = commit
	}

	service.Status.Source = instance.Spec.Source
	service.Status.Runtime = serverlessv1alpha1.RuntimeExtended(instance.Spec.Runtime)

	if err := r.client.Status().Update(ctx, service); err != nil {
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

func (r *FunctionReconciler) equalRepositories(existing serverlessv1alpha1.Repository, new *serverlessv1alpha1.Repository) bool {
	if new == nil {
		return true
	}
	expected := *new

	return existing.Reference == expected.Reference &&
		existing.BaseDir == expected.BaseDir
}

func (r *FunctionReconciler) getConditionStatus(conditions []serverlessv1alpha1.Condition, conditionType serverlessv1alpha1.ConditionType) corev1.ConditionStatus {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status
		}
	}

	return corev1.ConditionUnknown
}

func (r *FunctionReconciler) getConditionReason(conditions []serverlessv1alpha1.Condition, conditionType serverlessv1alpha1.ConditionType) serverlessv1alpha1.ConditionReason {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Reason
		}
	}

	return ""
}
