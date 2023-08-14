package serverless

import (
	"crypto/sha256"
	"fmt"
	"path"
	"reflect"
	"sort"
	"strings"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	FunctionSourceKey = "source"
	FunctionDepsKey   = "dependencies"
)

func getConditionStatus(conditions []serverlessv1alpha2.Condition, conditionType serverlessv1alpha2.ConditionType) corev1.ConditionStatus {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status
		}
	}

	return corev1.ConditionUnknown
}

func updateCondition(conditions []serverlessv1alpha2.Condition, condition serverlessv1alpha2.Condition) []serverlessv1alpha2.Condition {
	conditionTypes := make(map[serverlessv1alpha2.ConditionType]interface{}, 3)
	var result []serverlessv1alpha2.Condition

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

func equalConditions(existing, expected []serverlessv1alpha2.Condition) bool {
	if len(existing) != len(expected) {
		return false
	}

	existingMap := make(map[serverlessv1alpha2.ConditionType]serverlessv1alpha2.Condition, len(existing))
	for _, value := range existing {
		existingMap[value.Type] = value
	}

	for _, expectedCondition := range expected {
		existingCondition := existingMap[expectedCondition.Type]
		if !existingCondition.Equal(&expectedCondition) {
			return false
		}
	}
	return true
}

func equalRepositories(existing serverlessv1alpha2.Repository, new *serverlessv1alpha2.Repository) bool {
	if new == nil {
		return true
	}
	expected := *new

	return existing.Reference == expected.Reference &&
		existing.BaseDir == expected.BaseDir
}

func equalFunctionStatus(left, right serverlessv1alpha2.FunctionStatus) bool {
	if !equalConditions(left.Conditions, right.Conditions) {
		return false
	}

	if left.Repository != right.Repository ||
		left.Commit != right.Commit ||
		left.Runtime != right.Runtime {
		return false
	}
	return true
}

func equalJobs(existing batchv1.Job, expected batchv1.Job) bool {
	existingArgs := existing.Spec.Template.Spec.Containers[0].Args
	expectedArgs := expected.Spec.Template.Spec.Containers[0].Args

	// Compare destination argument as it contains fnImage tag
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

func getPackageConfigVolumeMountsForRuntime(rtm serverlessv1alpha2.Runtime) []corev1.VolumeMount {
	switch rtm {
	case serverlessv1alpha2.NodeJs16, serverlessv1alpha2.NodeJs18:
		return []corev1.VolumeMount{
			{
				Name:      "registry-config",
				ReadOnly:  true,
				MountPath: path.Join(workspaceMountPath, "registry-config/.npmrc"),
				SubPath:   ".npmrc",
			},
		}
	case serverlessv1alpha2.Python39:
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

func buildDeploymentEnvs(namespace, traceCollectorEndpoint, publisherProxyAddress string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "SERVICE_NAMESPACE", Value: namespace},
		{Name: "TRACE_COLLECTOR_ENDPOINT", Value: traceCollectorEndpoint},
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

// TODO refactor to make this code more readable
func equalDeployments(existing appsv1.Deployment, expected appsv1.Deployment) bool {
	return len(existing.Spec.Template.Spec.Containers) == 1 &&
		len(existing.Spec.Template.Spec.Containers) == len(expected.Spec.Template.Spec.Containers) &&
		existing.Spec.Template.Spec.Containers[0].Image == expected.Spec.Template.Spec.Containers[0].Image &&
		envsEqual(existing.Spec.Template.Spec.Containers[0].Env, expected.Spec.Template.Spec.Containers[0].Env) &&
		mapsEqual(existing.GetLabels(), expected.GetLabels()) &&
		mapsEqual(existing.Spec.Template.GetLabels(), expected.Spec.Template.GetLabels()) &&
		equalResources(existing.Spec.Template.Spec.Containers[0].Resources, expected.Spec.Template.Spec.Containers[0].Resources) &&
		equalInt32Pointer(existing.Spec.Replicas, expected.Spec.Replicas) &&
		equalSecretMounts(existing.Spec.Template.Spec, expected.Spec.Template.Spec) &&
		mapsEqual(existing.Spec.Template.GetAnnotations(), expected.Spec.Template.GetAnnotations())
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

func equalResources(existing, expected corev1.ResourceRequirements) bool {
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

func equalSecretMounts(existing, expected corev1.PodSpec) bool {
	existingSecretVolumes := filterOnlySecretVolumes(existing.Volumes)
	expectedSecretVolumes := filterOnlySecretVolumes(expected.Volumes)
	if !equalSecretVolumes(existingSecretVolumes, expectedSecretVolumes) {
		return false
	}

	existingSecretVolumeMounts := filterOnlyKnownVolumes(existing.Containers[0].VolumeMounts, existingSecretVolumes)
	expectedSecretVolumeMounts := filterOnlyKnownVolumes(expected.Containers[0].VolumeMounts, expectedSecretVolumes)
	return equalVolumeMounts(existingSecretVolumeMounts, expectedSecretVolumeMounts)
}

type secretVolumeMountSorter []corev1.VolumeMount

func (s secretVolumeMountSorter) Len() int           { return len(s) }
func (s secretVolumeMountSorter) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s secretVolumeMountSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func equalVolumeMounts(existing []corev1.VolumeMount, expected []corev1.VolumeMount) bool {
	sort.Stable(secretVolumeMountSorter(existing))
	sort.Stable(secretVolumeMountSorter(expected))
	return reflect.DeepEqual(existing, expected)
}

type secretVolumeSorter []corev1.Volume

func (s secretVolumeSorter) Len() int           { return len(s) }
func (s secretVolumeSorter) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s secretVolumeSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func equalSecretVolumes(existing []corev1.Volume, expected []corev1.Volume) bool {
	sort.Stable(secretVolumeSorter(existing))
	sort.Stable(secretVolumeSorter(expected))
	return reflect.DeepEqual(existing, expected)
}

func filterOnlyKnownVolumes(mounts []corev1.VolumeMount, knownVolumes []corev1.Volume) []corev1.VolumeMount {
	knownVolumeNames := getVolumeNames(knownVolumes)
	var knownVolumeMounts []corev1.VolumeMount
	for _, mount := range mounts {
		if stringInSlice(mount.Name, knownVolumeNames) {
			knownVolumeMounts = append(knownVolumeMounts, mount)
		}
	}
	return knownVolumeMounts
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func getVolumeNames(volumes []corev1.Volume) []string {
	var names []string
	for _, volume := range volumes {
		names = append(names, volume.Name)
	}
	return names
}

func filterOnlySecretVolumes(volumes []corev1.Volume) []corev1.Volume {
	var secretVolumes []corev1.Volume
	for _, volume := range volumes {
		if volume.Secret != nil {
			secretVolumes = append(secretVolumes, volume)
		}
	}
	return secretVolumes
}

func isScalingEnabled(instance *serverlessv1alpha2.Function) bool {
	if instance.Spec.ScaleConfig == nil {
		return false
	}
	return !equalInt32Pointer(instance.Spec.ScaleConfig.MinReplicas, instance.Spec.ScaleConfig.MaxReplicas)
}

func getConditionReason(conditions []serverlessv1alpha2.Condition, conditionType serverlessv1alpha2.ConditionType) serverlessv1alpha2.ConditionReason {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Reason
		}
	}

	return ""
}

func calculateInlineImageTag(instance *serverlessv1alpha2.Function) string {
	hash := sha256.Sum256([]byte(strings.Join([]string{
		string(instance.GetUID()),
		fmt.Sprintf("%v", *instance.Spec.Source.Inline),
		instance.EffectiveRuntime(),
	}, "-")))

	return fmt.Sprintf("%x", hash)
}

func calculateGitImageTag(instance *serverlessv1alpha2.Function) string {
	data := strings.Join([]string{
		string(instance.GetUID()),
		instance.Status.Commit,
		instance.Status.BaseDir,
		instance.EffectiveRuntime(),
	}, "-")
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
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
