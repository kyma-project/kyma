package serverless

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"

	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

func assertSuccessfulFunctionBuild(t *testing.T, resourceClient resource.Client, reconciler *FunctionReconciler, request ctrl.Request, fnLabels map[string]string, rebuilding bool) {
	g := gomega.NewGomegaWithT(t)

	initialDeploymentCondition := corev1.ConditionUnknown
	initialConditionsCount := 2
	if rebuilding {
		initialDeploymentCondition = corev1.ConditionTrue
		initialConditionsCount = 3
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Log("creating the Job")
	result, err := reconciler.Reconcile(ctx, request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 1))

	function := &serverlessv1alpha2.Function{}
	g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	g.Expect(function.Status.Conditions).To(gomega.HaveLen(initialConditionsCount))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(initialDeploymentCondition))

	jobList := &batchv1.JobList{}
	err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(jobList.Items).To(gomega.HaveLen(1))

	t.Log("build in progress")
	result, err = reconciler.Reconcile(ctx, request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 1))

	function = &serverlessv1alpha2.Function{}
	g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	g.Expect(function.Status.Conditions).To(gomega.HaveLen(initialConditionsCount))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(initialDeploymentCondition))

	g.Expect(getConditionReason(function.Status.Conditions, serverlessv1alpha2.ConditionBuildReady)).To(gomega.Equal(serverlessv1alpha2.ConditionReasonJobRunning))

	t.Log("build finished")
	job := &batchv1.Job{}
	g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
	g.Expect(job).ToNot(gomega.BeNil())
	job.Status.Succeeded = 1
	now := metav1.Now()
	job.Status.CompletionTime = &now
	g.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

	result, err = reconciler.Reconcile(ctx, request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 1))

	function = &serverlessv1alpha2.Function{}
	g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	g.Expect(function.Status.Conditions).To(gomega.HaveLen(initialConditionsCount))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(initialDeploymentCondition))

	g.Expect(getConditionReason(function.Status.Conditions, serverlessv1alpha2.ConditionBuildReady)).To(gomega.Equal(serverlessv1alpha2.ConditionReasonJobFinished))
}

func assertSuccessfulFunctionDeployment(t *testing.T, resourceClient resource.Client, reconciler *FunctionReconciler, request ctrl.Request, fnLabels map[string]string, regPullAddr string, redeployment bool) {
	g := gomega.NewGomegaWithT(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Log("deploy started")
	result, err := reconciler.Reconcile(ctx, request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 1))

	function := &serverlessv1alpha2.Function{}
	g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

	deployments := &appsv1.DeploymentList{}
	g.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
	g.Expect(len(deployments.Items)).To(gomega.Equal(1))
	deployment := &deployments.Items[0]
	g.Expect(deployment).ToNot(gomega.BeNil())
	g.Expect(deployment.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))

	s := systemState{
		// TODO https://github.com/kyma-project/kyma/issues/14079
		instance: *function,
	}

	g.Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(gomega.Equal(s.buildImageAddress(regPullAddr)))
	g.Expect(deployment.Spec.Template.Labels).To(gomega.HaveLen(7))
	g.Expect(deployment.Spec.Template.Labels[serverlessv1alpha2.FunctionNameLabel]).To(gomega.Equal(function.Name))
	g.Expect(deployment.Spec.Template.Labels[serverlessv1alpha2.FunctionManagedByLabel]).To(gomega.Equal(serverlessv1alpha2.FunctionControllerValue))
	g.Expect(deployment.Spec.Template.Labels[serverlessv1alpha2.FunctionUUIDLabel]).To(gomega.Equal(string(function.UID)))
	g.Expect(deployment.Spec.Template.Labels[serverlessv1alpha2.FunctionResourceLabel]).To(gomega.Equal(serverlessv1alpha2.FunctionResourceLabelDeploymentValue))
	g.Expect(deployment.Spec.Template.Labels[testBindingLabel1]).To(gomega.Equal("foobar"))
	g.Expect(deployment.Spec.Template.Labels[testBindingLabel2]).To(gomega.Equal(testBindingLabelValue))
	g.Expect(deployment.Spec.Template.Labels["foo"]).To(gomega.Equal("bar"))

	if !redeployment {
		t.Log("service creation")
		result, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 1))

		function = &serverlessv1alpha2.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		g.Expect(getConditionReason(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(serverlessv1alpha2.ConditionReasonServiceCreated))
	}

	t.Log("service ready")
	jobList := &batchv1.JobList{}
	err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(jobList.Items).To(gomega.HaveLen(1))
	job := &batchv1.Job{}
	g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
	g.Expect(job).ToNot(gomega.BeNil())

	svc := &corev1.Service{}
	g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, svc)).To(gomega.Succeed())
	g.Expect(err).To(gomega.BeNil())

	g.Expect(svc.Spec.Ports).To(gomega.HaveLen(1))
	g.Expect(svc.Spec.Ports[0].Name).To(gomega.Equal("http"))
	g.Expect(svc.Spec.Ports[0].TargetPort).To(gomega.Equal(intstr.FromInt(8080)))

	g.Expect(isSubset(svc.Spec.Selector, job.Spec.Template.Labels)).To(gomega.BeFalse(), "svc selector should not catch job pods")
	g.Expect(svc.Spec.Selector).To(gomega.Equal(deployment.Spec.Selector.MatchLabels))

	if !redeployment {
		t.Log("hpa creation")
		result, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 1))

		function = &serverlessv1alpha2.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		g.Expect(getConditionReason(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(serverlessv1alpha2.ConditionReasonHorizontalPodAutoscalerCreated))
	}

	t.Log("hpa ready")

	hpaList := &autoscalingv1.HorizontalPodAutoscalerList{}
	err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(hpaList.Items).To(gomega.HaveLen(1))

	hpaSpec := hpaList.Items[0].Spec

	g.Expect(hpaSpec.ScaleTargetRef.Name).To(gomega.Equal(function.GetName()))
	g.Expect(hpaSpec.ScaleTargetRef.Kind).To(gomega.Equal(serverlessv1alpha2.FunctionKind))
	g.Expect(hpaSpec.ScaleTargetRef.APIVersion).To(gomega.Equal(serverlessv1alpha2.GroupVersion.String()))

	t.Log("deployment ready")
	deployment.Status.Conditions = []appsv1.DeploymentCondition{
		{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: MinimumReplicasAvailable},
		{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: NewRSAvailableReason},
	}
	g.Expect(resourceClient.Status().Update(context.TODO(), deployment)).To(gomega.Succeed())

	result, err = reconciler.Reconcile(ctx, request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(time.Minute * 5))

	function = &serverlessv1alpha2.Function{}
	g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionReason(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(serverlessv1alpha2.ConditionReasonDeploymentReady))

	t.Log("should not change state on reconcile")
	result, err = reconciler.Reconcile(ctx, request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(time.Minute * 5))

	function = &serverlessv1alpha2.Function{}
	g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
	g.Expect(getConditionStatus(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))

	g.Expect(getConditionReason(function.Status.Conditions, serverlessv1alpha2.ConditionRunning)).To(gomega.Equal(serverlessv1alpha2.ConditionReasonDeploymentReady))
}
