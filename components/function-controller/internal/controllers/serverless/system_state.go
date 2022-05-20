package serverless

import (
	"fmt"
	"path"
	"strings"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	fnRuntime "github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/runtime"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SystemState interface{}

// TODO extract interface
type systemState struct {
	instance    serverlessv1alpha1.Function
	image       string               // TODO make sure this is needed
	configMaps  corev1.ConfigMapList // TODO create issue to refactor this (only 1 config map should be here)
	deployments appsv1.DeploymentList
	jobs        batchv1.JobList
	services    corev1.ServiceList
	hpas        autoscalingv1.HorizontalPodAutoscalerList
}

var _ SystemState = systemState{}

// TODO to self - create issue to refactor this
func (s *systemState) inlineSourceChanged(dockerPullAddress string) bool {
	image := s.instance.BuildImageAddress(dockerPullAddress)
	configurationStatus := getConditionStatus(s.instance.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)
	rtm := fnRuntime.GetRuntime(s.instance.Spec.Runtime)
	labels := s.instance.GetMergedLables()

	if len(s.deployments.Items) == 1 &&
		len(s.configMaps.Items) == 1 &&
		s.deployments.Items[0].Spec.Template.Spec.Containers[0].Image == image &&
		s.instance.Spec.Source == s.configMaps.Items[0].Data[FunctionSourceKey] &&
		rtm.SanitizeDependencies(s.instance.Spec.Deps) == s.configMaps.Items[0].Data[FunctionDepsKey] &&
		configurationStatus != corev1.ConditionUnknown &&
		mapsEqual(s.configMaps.Items[0].Labels, labels) {
		return false
	}

	return !(len(s.configMaps.Items) == 1 &&
		s.instance.Spec.Source == s.configMaps.Items[0].Data[FunctionSourceKey] &&
		rtm.SanitizeDependencies(s.instance.Spec.Deps) == s.configMaps.Items[0].Data[FunctionDepsKey] &&
		configurationStatus == corev1.ConditionTrue &&
		mapsEqual(s.configMaps.Items[0].Labels, labels))
}

func (s *systemState) buildJobChanged(expectedJob batchv1.Job) bool {
	conditionStatus := getConditionStatus(
		s.instance.Status.Conditions,
		serverlessv1alpha1.ConditionBuildReady,
	)

	if len(s.deployments.Items) == 1 &&
		s.deployments.Items[0].Spec.Template.Spec.Containers[0].Image == s.image &&
		conditionStatus != corev1.ConditionUnknown &&
		len(s.jobs.Items) > 0 &&
		mapsEqual(expectedJob.GetLabels(), s.jobs.Items[0].GetLabels()) {

		return conditionStatus == corev1.ConditionFalse
	}

	return len(s.jobs.Items) != 1 ||
		len(s.jobs.Items[0].Spec.Template.Spec.Containers) != 1 ||
		// Compare image argument
		!equalJobs(s.jobs.Items[0], expectedJob) ||
		!mapsEqual(expectedJob.GetLabels(), s.jobs.Items[0].GetLabels()) ||
		conditionStatus == corev1.ConditionUnknown ||
		conditionStatus == corev1.ConditionFalse
}

func (s *systemState) hpaEqual(targetCPUUtilizationPercentage int32) bool {
	if len(s.deployments.Items) == 0 {
		return false
	}

	expected := buildFunctionHPA(s.instance, s.deployments.Items[0].GetName(), targetCPUUtilizationPercentage)

	scalingEnabled := isScalingEnabled(&s.instance)

	numHpa := len(s.hpas.Items)
	return (scalingEnabled && numHpa != 1) ||
		(scalingEnabled && !equalHorizontalPodAutoscalers(s.hpas.Items[0], expected)) ||
		(!scalingEnabled && numHpa != 0)
}

func (s *systemState) defaultReplicas() (int32, int32) {
	min, max := int32(1), int32(1)
	spec := s.instance.Spec
	if spec.MinReplicas != nil && *spec.MinReplicas > 0 {
		min = *spec.MinReplicas
	}
	// special case
	if spec.MaxReplicas == nil || min > *spec.MaxReplicas {
		max = min
	} else {
		max = *spec.MaxReplicas
	}
	return min, max
}

func (s *systemState) buildHorizontalPodAutoscaler(targetCPUUtilizationPercentage int32) autoscalingv1.HorizontalPodAutoscaler {
	minReplicas, maxReplicas := s.defaultReplicas()
	deploymentName := s.deployments.Items[0].GetName()
	return autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", s.instance.GetName()),
			Namespace:    s.instance.GetNamespace(),
			Labels:       s.functionLabels(),
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       deploymentName,
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
			MinReplicas:                    &minReplicas,
			MaxReplicas:                    maxReplicas,
			TargetCPUUtilizationPercentage: &targetCPUUtilizationPercentage,
		},
	}
}

func (s *systemState) equalHorizontalPodAutoscalers(expected autoscalingv1.HorizontalPodAutoscaler) bool {
	existing := s.hpas.Items[0]
	return equalInt32Pointer(existing.Spec.TargetCPUUtilizationPercentage, expected.Spec.TargetCPUUtilizationPercentage) &&
		equalInt32Pointer(existing.Spec.MinReplicas, expected.Spec.MinReplicas) &&
		existing.Spec.MaxReplicas == expected.Spec.MaxReplicas &&
		mapsEqual(existing.Labels, expected.Labels) &&
		existing.Spec.ScaleTargetRef.Name == expected.Spec.ScaleTargetRef.Name
}

func (s *systemState) gitFnSrcChanged(commit string) bool {
	return s.instance.Status.Commit == "" ||
		commit != s.instance.Status.Commit ||
		s.instance.Spec.Reference != s.instance.Status.Reference ||
		serverlessv1alpha1.RuntimeExtended(s.instance.Spec.Runtime) != s.instance.Status.Runtime ||
		s.instance.Spec.BaseDir != s.instance.Status.BaseDir ||
		getConditionStatus(s.instance.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady) == corev1.ConditionFalse

}

func (s *systemState) getGitBuildJobVolumeMounts(rtmConfig runtime.Config) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{Name: "credentials", ReadOnly: true, MountPath: "/docker"},
		// Must be mounted with SubPath otherwise files are symlinks and it is not possible to use COPY in Dockerfile
		// If COPY is not used, then the cache will not work
		{Name: "workspace", MountPath: path.Join(workspaceMountPath, "src"), SubPath: strings.TrimPrefix(s.instance.Spec.BaseDir, "/")},
		{Name: "runtime", ReadOnly: true, MountPath: path.Join(workspaceMountPath, "Dockerfile"), SubPath: "Dockerfile"},
	}
	// add package registry config volume mount depending on the used runtime
	volumeMounts = append(volumeMounts, getPackageConfigVolumeMountsForRuntime(rtmConfig.Runtime)...)
	return volumeMounts
}

func (s *systemState) jobFailed(p func(reason string) bool) bool {
	if len(s.jobs.Items) == 0 {
		return false
	}

	return jobFailed(s.jobs.Items[0], p)
}
