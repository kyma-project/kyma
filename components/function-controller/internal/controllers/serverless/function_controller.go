package serverless

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

func (r *FunctionReconciler) updateStatusWithoutRepository(ctx context.Context, instance *serverlessv1alpha1.Function, condition serverlessv1alpha1.Condition) error {
	return r.updateStatus(ctx, instance, condition, nil, "")
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

func (r *FunctionReconciler) updateStatus(ctx context.Context, instance *serverlessv1alpha1.Function, condition serverlessv1alpha1.Condition, repository *serverlessv1alpha1.Repository, commit string) error {
	condition.LastTransitionTime = metav1.Now()
	currentFunction := &serverlessv1alpha1.Function{}
	err := r.client.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, currentFunction)
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	currentFunction.Status.Conditions = r.updateCondition(currentFunction.Status.Conditions, condition)

	equalConditions := r.equalConditions(instance.Status.Conditions, currentFunction.Status.Conditions)
	if equalConditions {
		if instance.Spec.Type != serverlessv1alpha1.SourceTypeGit {
			return nil
		}
		// checking if status changed in gitops flow
		if r.equalRepositories(instance.Status.Repository, repository) &&
			instance.Status.Commit == commit {
			return nil
		}
	}

	if repository != nil {
		currentFunction.Status.Repository = *repository
		currentFunction.Status.Commit = commit
	}
	currentFunction.Status.Source = instance.Spec.Source
	currentFunction.Status.Runtime = serverlessv1alpha1.RuntimeExtended(instance.Spec.Runtime)

	if !r.equalFunctionStatus(currentFunction.Status, instance.Status) {
		if err := r.client.Status().Update(ctx, currentFunction); err != nil {
			return err
		}

		r.statsCollector.UpdateReconcileStats(instance, condition)

		eventType := "Normal"
		if condition.Status == corev1.ConditionFalse {
			eventType = "Warning"
		}

		r.recorder.Event(currentFunction, eventType, string(condition.Reason), condition.Message)
	}
	return nil
}

func (r *FunctionReconciler) updateCondition(conditions []serverlessv1alpha1.Condition, condition serverlessv1alpha1.Condition) []serverlessv1alpha1.Condition {
	conditionTypes := make(map[serverlessv1alpha1.ConditionType]interface{}, 3)
	var result []serverlessv1alpha1.Condition

	result = append(result, condition)
	conditionTypes[condition.Type] = nil

	for _, value := range conditions {
		if _, ok := conditionTypes[value.Type]; !ok {
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

func (r *FunctionReconciler) equalFunctionStatus(left, right serverlessv1alpha1.FunctionStatus) bool {
	if !r.equalConditions(left.Conditions, right.Conditions) {
		return false
	}

	if left.Repository != right.Repository ||
		left.Commit != right.Commit ||
		left.Source != right.Source ||
		left.Runtime != right.Runtime {
		return false
	}
	return true
}
