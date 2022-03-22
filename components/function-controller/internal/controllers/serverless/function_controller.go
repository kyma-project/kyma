package serverless

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
)

const (
	FunctionSourceKey = "source"
	FunctionDepsKey   = "dependencies"
)

func mapsEqual(existing, expected map[string]string) bool {
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

func envsEqual(existing, expected []corev1.EnvVar) bool {
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

func (r *FunctionReconciler) updateStatusWithoutRepository(ctx context.Context, instance *serverlessv1alpha1.Function, condition serverlessv1alpha1.Condition) error {
	return r.updateStatus(ctx, instance, condition, nil, "")
}

func (r *FunctionReconciler) updateStatus(ctx context.Context, instance *serverlessv1alpha1.Function, condition serverlessv1alpha1.Condition, repository *serverlessv1alpha1.Repository, commit string) error {
	condition.LastTransitionTime = metav1.Now()
	currentFunction := &serverlessv1alpha1.Function{}
	err := r.client.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, currentFunction)
	if err != nil {
		return client.IgnoreNotFound(err)
	}

	currentFunction.Status.Conditions = updateCondition(currentFunction.Status.Conditions, condition)

	equalConditions := equalConditions(instance.Status.Conditions, currentFunction.Status.Conditions)
	if equalConditions {
		if instance.Spec.Type != serverlessv1alpha1.SourceTypeGit {
			return nil
		}
		// checking if status changed in gitops flow
		if equalRepositories(instance.Status.Repository, repository) &&
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

	if !equalFunctionStatus(currentFunction.Status, instance.Status) {
		if err := r.client.Status().Update(ctx, currentFunction); err != nil {
			return errors.Wrap(err, "while updating function status")
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

func updateCondition(conditions []serverlessv1alpha1.Condition, condition serverlessv1alpha1.Condition) []serverlessv1alpha1.Condition {
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

func equalConditions(existing, expected []serverlessv1alpha1.Condition) bool {
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

func equalRepositories(existing serverlessv1alpha1.Repository, new *serverlessv1alpha1.Repository) bool {
	if new == nil {
		return true
	}
	expected := *new

	return existing.Reference == expected.Reference &&
		existing.BaseDir == expected.BaseDir
}

func getConditionStatus(conditions []serverlessv1alpha1.Condition, conditionType serverlessv1alpha1.ConditionType) corev1.ConditionStatus {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status
		}
	}

	return corev1.ConditionUnknown
}

func getConditionReason(conditions []serverlessv1alpha1.Condition, conditionType serverlessv1alpha1.ConditionType) serverlessv1alpha1.ConditionReason {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Reason
		}
	}

	return ""
}

func equalFunctionStatus(left, right serverlessv1alpha1.FunctionStatus) bool {
	if !equalConditions(left.Conditions, right.Conditions) {
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

func readDockerConfig(ctx context.Context, client resource.Client, config FunctionConfig, instance *serverlessv1alpha1.Function) (DockerConfig, error) {
	var secret corev1.Secret
	// try reading user config
	if err := client.Get(ctx, ctrlclient.ObjectKey{Namespace: instance.Namespace, Name: config.ImageRegistryExternalDockerConfigSecretName}, &secret); err == nil {
		data := readSecretData(secret.Data)
		return DockerConfig{
			ActiveRegistryConfigSecretName: config.ImageRegistryExternalDockerConfigSecretName,
			PushAddress:                    data["registryAddress"],
			PullAddress:                    data["registryAddress"],
		}, nil
	}

	// try reading default config
	if err := client.Get(ctx, ctrlclient.ObjectKey{Namespace: instance.Namespace, Name: config.ImageRegistryDefaultDockerConfigSecretName}, &secret); err == nil {
		data := readSecretData(secret.Data)
		if data["isInternal"] == "true" {
			return DockerConfig{
				ActiveRegistryConfigSecretName: config.ImageRegistryDefaultDockerConfigSecretName,
				PushAddress:                    data["registryAddress"],
				PullAddress:                    data["serverAddress"],
			}, nil
		} else {
			return DockerConfig{
				ActiveRegistryConfigSecretName: config.ImageRegistryDefaultDockerConfigSecretName,
				PushAddress:                    data["registryAddress"],
				PullAddress:                    data["registryAddress"],
			}, nil
		}
	}

	return DockerConfig{}, errors.Errorf("Docker registry configuration not found, none of configuration secrets (%s, %s) found in function namespace", config.ImageRegistryDefaultDockerConfigSecretName, config.ImageRegistryExternalDockerConfigSecretName)
}

func readSecretData(data map[string][]byte) map[string]string {
	output := make(map[string]string)
	for k, v := range data {
		output[k] = string(v)
	}
	return output
}
