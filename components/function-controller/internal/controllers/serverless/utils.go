package serverless

import (
	"path"
	"strings"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	FunctionSourceKey = "source"
	FunctionDepsKey   = "dependencies"
)

func getConditionStatus(conditions []serverlessv1alpha1.Condition, conditionType serverlessv1alpha1.ConditionType) corev1.ConditionStatus {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status
		}
	}

	return corev1.ConditionUnknown
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

func equalJobs(existing batchv1.Job, expected batchv1.Job) bool {
	existingArgs := existing.Spec.Template.Spec.Containers[0].Args
	expectedArgs := expected.Spec.Template.Spec.Containers[0].Args

	// Compare destination argument as it contains image tag
	existingDst := getArg(existingArgs, destinationArg)
	expectedDst := getArg(expectedArgs, destinationArg)

	return existingDst == expectedDst
}

func getArg(args []string, arg string) string {
	for _, item := range args {
		if strings.HasPrefix(item, arg) {
			return item
		}
	}
	return ""
}

func getBuildJobVolumeMounts(rtmConfig runtime.Config) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
		// If COPY is not used, then the cache will not work
		{Name: "sources", ReadOnly: true, MountPath: path.Join(baseDir, rtmConfig.DependencyFile), SubPath: FunctionDepsKey},
		{Name: "sources", ReadOnly: true, MountPath: path.Join(baseDir, rtmConfig.FunctionFile), SubPath: FunctionSourceKey},
		{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
		{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
	}
	// add package registry config volume mount depending on the used runtime
	volumeMounts = append(volumeMounts, getPackageConfigVolumeMountsForRuntime(rtmConfig.Runtime)...)
	return volumeMounts
}

func getPackageConfigVolumeMountsForRuntime(rtm serverlessv1alpha1.Runtime) []corev1.VolumeMount {
	switch rtm {
	case serverlessv1alpha1.Nodejs12, serverlessv1alpha1.Nodejs14:
		return []corev1.VolumeMount{
			{
				Name:      "registry-config",
				ReadOnly:  true,
				MountPath: path.Join(workspaceMountPath, "registry-config/.npmrc"),
				SubPath:   ".npmrc",
			},
		}
	case serverlessv1alpha1.Python39:
		return []corev1.VolumeMount{
			{
				Name:      "registry-config",
				ReadOnly:  true,
				MountPath: path.Join(workspaceMountPath, "registry-config/pip.conf"),
				SubPath:   "pip.conf"},
		}
	}
	return nil
}

func didNotSucceed(j batchv1.Job) bool {
	return j.Status.Succeeded == 0
}

func didNotFail(j batchv1.Job) bool {
	return j.Status.Failed == 0
}

func countJobs(l batchv1.JobList, predicates ...func(batchv1.Job) bool) int {
	var out int

processing_next_item:
	for _, j := range l.Items {
		for _, p := range predicates {
			if !p(j) {
				continue processing_next_item
			}
		}
		out++
	}

	return out
}

func buildDeploymentEnvs(namespace, jaegerServiceEndpoint, publisherProxyAddress string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "SERVICE_NAMESPACE", Value: namespace},
		{Name: "JAEGER_SERVICE_ENDPOINT", Value: jaegerServiceEndpoint},
		{Name: "PUBLISHER_PROXY_ADDRESS", Value: publisherProxyAddress},
		{Name: "FUNC_HANDLER", Value: "main"},
		{Name: "MOD_NAME", Value: "handler"},
		{Name: "FUNC_PORT", Value: "8080"},
	}
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

func containersEqual(existing, expected []corev1.Container) bool {
	if len(existing) != 1 || len(expected) != 1 {
		return false
	}
	if existing[0].Image != expected[0].Image {
		return false
	}
	if !envsEqual(existing[0].Env, expected[0].Env) {
		return false
	}
	if !resourcesEqual(existing[0].Resources, expected[0].Resources) {
		return false
	}

	return true
}

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

//TODO refactor to make this code more readable
func equalDeployments(existing appsv1.Deployment, expected appsv1.Deployment, scalingEnabled bool) bool {
	equalContainerConfig := containersEqual(existing.Spec.Template.Spec.Containers, expected.Spec.Template.Spec.Containers)

	equalLabels := mapsEqual(existing.GetLabels(), expected.GetLabels()) &&
		mapsEqual(existing.Spec.Template.GetLabels(), expected.Spec.Template.GetLabels())
	equalScalingConfig := scalingEnabled || equalInt32Pointer(existing.Spec.Replicas, expected.Spec.Replicas)

	return equalContainerConfig && equalLabels && equalScalingConfig
}

func equalServices(existing corev1.Service, expected corev1.Service) bool {
	return mapsEqual(existing.Spec.Selector, expected.Spec.Selector) &&
		mapsEqual(existing.Labels, expected.Labels) &&
		len(existing.Spec.Ports) == len(expected.Spec.Ports) &&
		len(expected.Spec.Ports) > 0 &&
		len(existing.Spec.Ports) > 0 &&
		existing.Spec.Ports[0].String() == expected.Spec.Ports[0].String()
}

func readSecretData(data map[string][]byte) map[string]string {
	output := make(map[string]string)
	for k, v := range data {
		output[k] = string(v)
	}
	return output
}

func resourcesEqual(existing, expected corev1.ResourceRequirements) bool {
	return existing.Requests.Memory().Equal(*expected.Requests.Memory()) &&
		existing.Requests.Cpu().Equal(*expected.Requests.Cpu()) &&
		existing.Limits.Memory().Equal(*expected.Limits.Memory()) &&
		existing.Limits.Cpu().Equal(*expected.Limits.Cpu())
}

func equalInt32Pointer(first *int32, second *int32) bool {
	if first == nil && second == nil {
		return true
	}
	if (first != nil && second == nil) || (first == nil && second != nil) {
		return false
	}

	return *first == *second
}

func isScalingEnabled(instance *serverlessv1alpha1.Function) bool {
	return !equalInt32Pointer(instance.Spec.MinReplicas, instance.Spec.MaxReplicas)
}

func getConditionReason(conditions []serverlessv1alpha1.Condition, conditionType serverlessv1alpha1.ConditionType) serverlessv1alpha1.ConditionReason {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Reason
		}
	}

	return ""
}

func equalHorizontalPodAutoscalers(existing, expected autoscalingv1.HorizontalPodAutoscaler) bool {
	equalCPUPercentage := equalInt32Pointer(existing.Spec.TargetCPUUtilizationPercentage, expected.Spec.TargetCPUUtilizationPercentage)
	equalReplicas := equalInt32Pointer(existing.Spec.MinReplicas, expected.Spec.MinReplicas) &&
		existing.Spec.MaxReplicas == expected.Spec.MaxReplicas
	equalLabels := mapsEqual(existing.Labels, expected.Labels)
	equalTargetName := existing.Spec.ScaleTargetRef.Name == expected.Spec.ScaleTargetRef.Name

	return equalCPUPercentage && equalReplicas &&
		equalLabels && equalTargetName
}

func jobFailed(job batchv1.Job, p func(reason string) bool) bool {
	for _, condition := range job.Status.Conditions {
		isFailedType := condition.Type == batchv1.JobFailed
		isStatusTrue := condition.Status == corev1.ConditionTrue

		if isFailedType && isStatusTrue {
			return p(condition.Reason)
		}
	}

	return false
}
