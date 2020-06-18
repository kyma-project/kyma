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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const (
	testBindingLabel1     = "use-ec7cd950-9c2b-45a4-9f63-556fd8ea07f4"
	testBindingLabel2     = "use-ec7cd950-9c2b-45a4-9f63-556fd8ea07f5"
	testBindingLabelValue = "146000"
	conditionLen          = 3
)

var _ = ginkgo.Describe("Function", func() {
	var (
		reconciler *FunctionReconciler
		request    ctrl.Request
	)

	ginkgo.BeforeEach(func() {
		function := newFixFunction("tutaj", "ah-tak-przeciez", 1, 2)
		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: function.GetNamespace(), Name: function.GetName()}}
		gomega.Expect(resourceClient.Create(context.TODO(), function)).To(gomega.Succeed())

		reconciler = NewFunction(resourceClient, log.Log, config, record.NewFakeRecorder(100))
	})

	ginkgo.It("should successfully create Function", func() {
		ginkgo.By("creating the ConfigMap")
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function := &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(1))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonConfigMapCreated))

		configMapList := &corev1.ConfigMapList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), configMapList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(configMapList.Items).To(gomega.HaveLen(1))
		gomega.Expect(configMapList.Items[0].Data[configMapFunction]).To(gomega.Equal(function.Spec.Source))
		gomega.Expect(configMapList.Items[0].Data[configMapHandler]).To(gomega.Equal("handler.main"))
		gomega.Expect(configMapList.Items[0].Data[configMapDeps]).To(gomega.Equal("{}"))

		ginkgo.By("creating the Job")
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(2))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		jobList := &batchv1.JobList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), jobList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(jobList.Items).To(gomega.HaveLen(1))

		ginkgo.By("build in progress")
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(2))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

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
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(2))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonJobFinished))

		ginkgo.By("deploy started")
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		deployments := &appsv1.DeploymentList{}
		gomega.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, reconciler.internalFunctionLabels(function), deployments)).To(gomega.Succeed())
		gomega.Expect(len(deployments.Items)).To(gomega.Equal(1))
		deployment := &deployments.Items[0]
		gomega.Expect(deployment).ToNot(gomega.BeNil())
		gomega.Expect(deployment.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(gomega.Equal(reconciler.buildExternalImageAddress(function)))
		gomega.Expect(deployment.Spec.Template.Labels).To(gomega.HaveLen(6)) // function-name, managed-by, uuid + 3
		gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionNameLabel]).To(gomega.Equal(function.Name))
		gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionManagedByLabel]).To(gomega.Equal("function-controller"))
		gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionUUIDLabel]).To(gomega.Equal(string(function.UID)))
		gomega.Expect(deployment.Spec.Template.Labels[testBindingLabel1]).To(gomega.Equal("foobar"))
		gomega.Expect(deployment.Spec.Template.Labels[testBindingLabel2]).To(gomega.Equal(testBindingLabelValue))
		gomega.Expect(deployment.Spec.Template.Labels["foo"]).To(gomega.Equal("bar"))

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

		svc := &corev1.Service{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, svc)).To(gomega.Succeed())
		gomega.Expect(err).To(gomega.BeNil())

		gomega.Expect(svc.Spec.Ports).To(gomega.HaveLen(1))
		gomega.Expect(svc.Spec.Ports[0].Name).To(gomega.Equal("http"))
		gomega.Expect(svc.Spec.Ports[0].TargetPort).To(gomega.Equal(intstr.FromInt(8080)))

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

		hpaList := &autoscalingv1.HorizontalPodAutoscalerList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), hpaList)
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
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Duration(0)))

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
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Duration(0)))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))

		gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentReady))
	})

	ginkgo.It("should set proper status on deployment fail", func() {
		ginkgo.By("creating cm")
		_, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		ginkgo.By("creating job")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		function := &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())

		jobList := &batchv1.JobList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), jobList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(jobList.Items).To(gomega.HaveLen(1))

		job := &batchv1.Job{}
		gomega.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
		gomega.Expect(job).ToNot(gomega.BeNil())
		job.Status.Succeeded = 1
		now := metav1.Now()
		job.Status.CompletionTime = &now
		gomega.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

		ginkgo.By("job finished")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		ginkgo.By("creating deployment")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		ginkgo.By("creating svc")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		ginkgo.By("creating hpa")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		ginkgo.By("deployment failed")
		deployments := &appsv1.DeploymentList{}
		gomega.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, reconciler.internalFunctionLabels(function), deployments)).To(gomega.Succeed())
		gomega.Expect(len(deployments.Items)).To(gomega.Equal(1))
		deployment := &deployments.Items[0]
		deployment.Status.Conditions = []appsv1.DeploymentCondition{
			{Type: appsv1.DeploymentReplicaFailure, Status: corev1.ConditionTrue, Message: "Some random message", Reason: "some reason"},
			{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionFalse, Message: "Deployment doesn't have minimum availability."},
		}

		gomega.Expect(resourceClient.Status().Update(context.TODO(), deployment)).To(gomega.Succeed())

		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionFalse))

		gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentFailed))
	})

	ginkgo.It("should properly handle apiserver lags, when two resources are created by accident", func() {
		ginkgo.By("creating cm")
		_, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		function := &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())

		configMapList := &corev1.ConfigMapList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), configMapList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(configMapList.Items).To(gomega.HaveLen(1))

		cm := configMapList.Items[0].DeepCopy()
		cm.Name = "" // generateName will create this
		cm.ResourceVersion = ""
		cm.UID = ""
		cm.CreationTimestamp = metav1.Time{}
		gomega.Expect(resourceClient.Create(context.TODO(), cm)).To(gomega.Succeed())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), configMapList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(configMapList.Items).To(gomega.HaveLen(2))

		ginkgo.By("deleting all configMaps")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), configMapList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(configMapList.Items).To(gomega.HaveLen(0))

		ginkgo.By("creating configMap again")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), configMapList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(configMapList.Items).To(gomega.HaveLen(1))

		ginkgo.By("creating job")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())

		jobList := &batchv1.JobList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), jobList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(jobList.Items).To(gomega.HaveLen(1))

		ginkgo.By("accidently creating second job -> apiserver lag")

		excessJob := jobList.Items[0].DeepCopy()
		excessJob.Name = "" // generateName will create this
		excessJob.ResourceVersion = ""
		excessJob.UID = ""
		excessJob.CreationTimestamp = metav1.Time{}
		excessJob.Spec.Selector = nil
		delete(excessJob.Spec.Template.ObjectMeta.Labels, "controller-uid")
		delete(excessJob.Spec.Template.ObjectMeta.Labels, "job-name")
		gomega.Expect(resourceClient.Create(context.TODO(), excessJob)).To(gomega.Succeed())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), jobList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(jobList.Items).To(gomega.HaveLen(2))

		ginkgo.By("deleting all jobs")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), jobList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(jobList.Items).To(gomega.HaveLen(0))

		ginkgo.By("creating job again")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), jobList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(jobList.Items).To(gomega.HaveLen(1))
		job := &batchv1.Job{}
		gomega.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
		gomega.Expect(job).ToNot(gomega.BeNil())
		job.Status.Succeeded = 1
		now := metav1.Now()
		job.Status.CompletionTime = &now
		gomega.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

		ginkgo.By("job finished")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		ginkgo.By("creating deployment")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		deployList := &appsv1.DeploymentList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.internalFunctionLabels(function), deployList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(deployList.Items).To(gomega.HaveLen(1))

		ginkgo.By("creating next deployment by accident")

		excessDeploy := deployList.Items[0].DeepCopy()
		excessDeploy.Name = "" // generateName will create this
		excessDeploy.ResourceVersion = ""
		excessDeploy.UID = ""
		excessDeploy.CreationTimestamp = metav1.Time{}

		ginkgo.By("creating svc")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		ginkgo.By("creating hpa")
		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())

		ginkgo.By("deployment ready")
		deployments := &appsv1.DeploymentList{}
		gomega.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, reconciler.internalFunctionLabels(function), deployments)).To(gomega.Succeed())
		gomega.Expect(len(deployments.Items)).To(gomega.Equal(1))
		deployment := &deployments.Items[0]

		deployment.Status.Conditions = []appsv1.DeploymentCondition{
			{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: MinimumReplicasAvailable},
			{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: NewRSAvailableReason},
		}
		gomega.Expect(resourceClient.Status().Update(context.TODO(), deployment)).To(gomega.Succeed())

		_, err = reconciler.Reconcile(request)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("should handle reconcilation lags", func() {
		ginkgo.By("handling not existing Function")
		result, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "nope", Name: "noooooopppeee"}})
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))
	})
})
