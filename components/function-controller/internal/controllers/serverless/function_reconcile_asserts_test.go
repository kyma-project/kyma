package serverless

import (
	"context"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func assertSuccessfulFunctionBuild(reconciler *FunctionReconciler, request ctrl.Request, fnLabels map[string]string, rebuilding bool) {
	initialDeploymentCondition := corev1.ConditionUnknown
	initialConditionsCount := 2
	if rebuilding {
		initialDeploymentCondition = corev1.ConditionTrue
		initialConditionsCount = 3
	}

	ginkgo.By("creating the Job")
	result, err := reconciler.Reconcile(request)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result.Requeue).To(gomega.BeFalse())
	gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

	function := &serverlessv1alpha1.Function{}
	gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(initialConditionsCount))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(initialDeploymentCondition))

	jobList := &batchv1.JobList{}
	err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(jobList.Items).To(gomega.HaveLen(1))

	ginkgo.By("build in progress")
	result, err = reconciler.Reconcile(request)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result.Requeue).To(gomega.BeFalse())
	gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

	function = &serverlessv1alpha1.Function{}
	gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(initialConditionsCount))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(initialDeploymentCondition))

	gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonJobRunning))

	ginkgo.By("build finished")
	job := &batchv1.Job{}
	gomega.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
	gomega.Expect(job).ToNot(gomega.BeNil())
	job.Status.Succeeded = 1
	now := metav1.Now()
	job.Status.CompletionTime = &now
	gomega.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

	result, err = reconciler.Reconcile(request)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result.Requeue).To(gomega.BeFalse())
	gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

	function = &serverlessv1alpha1.Function{}
	gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(initialConditionsCount))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(initialDeploymentCondition))

	gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonJobFinished))
}

func assertSuccessfulFunctionDeployment(reconciler *FunctionReconciler, request ctrl.Request, fnLabels map[string]string, registryAddress string, redeployment bool) {
	ginkgo.By("deploy started")
	result, err := reconciler.Reconcile(request)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result.Requeue).To(gomega.BeFalse())
	gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

	function := &serverlessv1alpha1.Function{}
	gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

	deployments := &appsv1.DeploymentList{}
	gomega.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
	gomega.Expect(len(deployments.Items)).To(gomega.Equal(1))
	deployment := &deployments.Items[0]
	gomega.Expect(deployment).ToNot(gomega.BeNil())
	gomega.Expect(deployment.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
	gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(gomega.Equal(reconciler.buildImageAddress(function, registryAddress)))
	gomega.Expect(deployment.Spec.Template.Labels).To(gomega.HaveLen(7))
	gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionNameLabel]).To(gomega.Equal(function.Name))
	gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionManagedByLabel]).To(gomega.Equal(serverlessv1alpha1.FunctionControllerValue))
	gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionUUIDLabel]).To(gomega.Equal(string(function.UID)))
	gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionResourceLabel]).To(gomega.Equal(serverlessv1alpha1.FunctionResourceLabelDeploymentValue))
	gomega.Expect(deployment.Spec.Template.Labels[testBindingLabel1]).To(gomega.Equal("foobar"))
	gomega.Expect(deployment.Spec.Template.Labels[testBindingLabel2]).To(gomega.Equal(testBindingLabelValue))
	gomega.Expect(deployment.Spec.Template.Labels["foo"]).To(gomega.Equal("bar"))

	if !redeployment {
		ginkgo.By("service creation")
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Duration(0)))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonServiceCreated))
	}

	ginkgo.By("service ready")
	jobList := &batchv1.JobList{}
	err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(jobList.Items).To(gomega.HaveLen(1))
	job := &batchv1.Job{}
	gomega.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
	gomega.Expect(job).ToNot(gomega.BeNil())

	svc := &corev1.Service{}
	gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, svc)).To(gomega.Succeed())
	gomega.Expect(err).To(gomega.BeNil())

	gomega.Expect(svc.Spec.Ports).To(gomega.HaveLen(1))
	gomega.Expect(svc.Spec.Ports[0].Name).To(gomega.Equal("http"))
	gomega.Expect(svc.Spec.Ports[0].TargetPort).To(gomega.Equal(intstr.FromInt(8080)))

	gomega.Expect(labels.AreLabelsInWhiteList(svc.Spec.Selector, job.Spec.Template.Labels)).To(gomega.BeFalse(), "svc selector should not catch job pods")
	gomega.Expect(svc.Spec.Selector).To(gomega.Equal(deployment.Spec.Selector.MatchLabels))

	if !redeployment {
		ginkgo.By("hpa creation")
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Duration(0)))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonHorizontalPodAutoscalerCreated))
	}

	ginkgo.By("hpa ready")

	hpaList := &autoscalingv1.HorizontalPodAutoscalerList{}
	err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(hpaList.Items).To(gomega.HaveLen(1))

	hpaSpec := hpaList.Items[0].Spec

	gomega.Expect(hpaSpec.ScaleTargetRef.Name).To(gomega.Equal(deployment.GetName()))
	gomega.Expect(hpaSpec.ScaleTargetRef.Kind).To(gomega.Equal("Deployment"))
	gomega.Expect(hpaSpec.ScaleTargetRef.APIVersion).To(gomega.Equal(appsv1.SchemeGroupVersion.String()))

	ginkgo.By("deployment ready")
	deployment.Status.Conditions = []appsv1.DeploymentCondition{
		{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: MinimumReplicasAvailable},
		{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: NewRSAvailableReason},
	}
	gomega.Expect(resourceClient.Status().Update(context.TODO(), deployment)).To(gomega.Succeed())

	result, err = reconciler.Reconcile(request)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result.Requeue).To(gomega.BeFalse())
	gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Minute * 5))

	function = &serverlessv1alpha1.Function{}
	gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentReady))

	ginkgo.By("should not change state on reconcile")
	result, err = reconciler.Reconcile(request)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result.Requeue).To(gomega.BeFalse())
	gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Minute * 5))

	function = &serverlessv1alpha1.Function{}
	gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
	gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
	gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))

	gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentReady))
}
