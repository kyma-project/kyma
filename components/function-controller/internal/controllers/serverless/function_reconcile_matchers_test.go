package serverless

import (
	"time"

	v1 "k8s.io/api/apps/v1"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	gtypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
)

var beOKReconcileResult = recResultMatcher(false, time.Second*1)
var beFinishedReconcileResult = recResultMatcher(false, time.Minute*5)

func recResultMatcher(requeue bool, requeueAfter time.Duration) gtypes.GomegaMatcher {
	return gstruct.MatchAllFields(gstruct.Fields{
		"RequeueAfter": gomega.Equal(requeueAfter),
		"Requeue":      gomega.Equal(requeue),
	})
}

var (
	haveConditionReasonSourceUpdated = haveConditionReason(
		serverlessv1alpha2.ConditionConfigurationReady,
		serverlessv1alpha2.ConditionReasonSourceUpdated,
	)

	haveConditionReasonJobRunning = haveConditionReason(
		serverlessv1alpha2.ConditionBuildReady,
		serverlessv1alpha2.ConditionReasonJobRunning)

	haveConditionReasonJobFinished = haveConditionReason(
		serverlessv1alpha2.ConditionBuildReady,
		serverlessv1alpha2.ConditionReasonJobFinished)

	haveConditionReasonServiceCreated = haveConditionReason(
		serverlessv1alpha2.ConditionRunning,
		serverlessv1alpha2.ConditionReasonServiceCreated)

	haveConditionReasonDeploymentReady = haveConditionReason(
		serverlessv1alpha2.ConditionRunning,
		serverlessv1alpha2.ConditionReasonDeploymentReady)

	haveConditionReasonHorizontalPodAutoscalerCreated = haveConditionReason(
		serverlessv1alpha2.ConditionRunning,
		serverlessv1alpha2.ConditionReasonHorizontalPodAutoscalerCreated)
)

func haveConditionReason(t serverlessv1alpha2.ConditionType, expected serverlessv1alpha2.ConditionReason) gtypes.GomegaMatcher {
	return gomega.WithTransform(func(fn *serverlessv1alpha2.Function) serverlessv1alpha2.ConditionReason {
		return getConditionReason(fn.Status.Conditions, t)
	}, gomega.Equal(expected))
}

var (
	haveConditionCfgRdy          = haveCondition(serverlessv1alpha2.ConditionConfigurationReady, corev1.ConditionTrue)
	haveConditionBuildRdy        = haveCondition(serverlessv1alpha2.ConditionBuildReady, corev1.ConditionTrue)
	haveUnknownConditionBuildRdy = haveCondition(serverlessv1alpha2.ConditionBuildReady, corev1.ConditionUnknown)
	haveConditionRunning         = haveCondition(serverlessv1alpha2.ConditionRunning, corev1.ConditionTrue)
	haveUnknownConditionRunning  = haveCondition(serverlessv1alpha2.ConditionRunning, corev1.ConditionUnknown)
)

func haveCondition(t serverlessv1alpha2.ConditionType, expected corev1.ConditionStatus) gtypes.GomegaMatcher {
	return gomega.WithTransform(func(fn *serverlessv1alpha2.Function) corev1.ConditionStatus {
		return getConditionStatus(fn.Status.Conditions, t)
	}, gomega.Equal(expected))
}

func haveConditionLen(expected int) gtypes.GomegaMatcher {
	return gomega.WithTransform(func(fn *serverlessv1alpha2.Function) int {
		return len(fn.Status.Conditions)
	}, gomega.Equal(expected))
}

func haveStatusReference(expected string) gtypes.GomegaMatcher {
	return gomega.WithTransform(func(fn *serverlessv1alpha2.Function) string {
		return fn.Status.Reference
	}, gomega.Equal(expected))
}

func haveStatusCommit(expected string) gtypes.GomegaMatcher {
	return gomega.WithTransform(func(fn *serverlessv1alpha2.Function) string {
		return fn.Status.Commit
	}, gomega.Equal(expected))
}

func haveSpecificContainer0Image(expected string) gtypes.GomegaMatcher {
	return gomega.And(
		gomega.WithTransform(func(d *v1.Deployment) int {
			return len(d.Spec.Template.Spec.Containers)
		}, gomega.BeNumerically(">=", 0)),
		gomega.WithTransform(func(d *v1.Deployment) string {
			return d.Spec.Template.Spec.Containers[0].Image
		}, gomega.Equal(expected)),
	)
}

func haveLabelWithValue(key, value interface{}) gtypes.GomegaMatcher {
	return gomega.WithTransform(func(d *v1.Deployment) map[string]string {
		return d.Spec.Template.Labels
	}, gomega.HaveKeyWithValue(key, value))
}

func haveLabelLen(expected int) gtypes.GomegaMatcher {
	return gomega.WithTransform(func(d *v1.Deployment) int {
		return len(d.Spec.Template.Labels)
	}, gomega.Equal(expected))
}
