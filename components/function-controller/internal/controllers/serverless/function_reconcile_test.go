package serverless

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/automock"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
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
	addedLabelKey         = "that-label"
	addedLabelValue       = "wasnt-here"
)

func TestFunctionReconciler_Reconcile(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	rtm := serverlessv1alpha1.Nodejs12
	resourceClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	testCfg := setUpControllerConfig(g)
	initializeServerlessResources(g, resourceClient)
	createDockerfileForRuntime(g, resourceClient, rtm)
	statsCollector := &automock.StatsCollector{}
	statsCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()

	reconciler := NewFunction(resourceClient, log.Log, testCfg, nil, record.NewFakeRecorder(100), statsCollector, make(chan bool))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("should successfully create Function", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		inFunction := newFixFunction(testNamespace, "success", 1, 2)
		g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())
		defer deleteFunction(g, resourceClient, inFunction)

		fnLabels := reconciler.functionLabels(inFunction)
		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: inFunction.GetNamespace(), Name: inFunction.GetName()}}

		//WHEN
		t.Log("creating the ConfigMap")
		result, err := reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function := &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(1))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonConfigMapCreated))

		configMapList := &corev1.ConfigMapList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, configMapList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(configMapList.Items).To(gomega.HaveLen(1))
		g.Expect(configMapList.Items[0].Data[FunctionSourceKey]).To(gomega.Equal(function.Spec.Source))
		g.Expect(configMapList.Items[0].Data[FunctionDepsKey]).To(gomega.Equal("{}"))

		assertSuccessfulFunctionBuild(t, resourceClient, reconciler, request, fnLabels, false)

		assertSuccessfulFunctionDeployment(t, resourceClient, reconciler, request, fnLabels, "registry.kyma.local", false)

		t.Log("should detect registry configuration change and rebuild function")
		customDockerRegistryConfiguration := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "serverless-registry-config",
				Namespace: testNamespace,
			},
			StringData: map[string]string{
				"registryAddress": "registry.external.host",
			},
		}
		g.Expect(resourceClient.Create(context.TODO(), &customDockerRegistryConfiguration)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Duration(0)))

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))

		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonJobsDeleted))

		assertSuccessfulFunctionBuild(t, resourceClient, reconciler, request, fnLabels, true)

		assertSuccessfulFunctionDeployment(t, resourceClient, reconciler, request, fnLabels, "registry.external.host", true)

		t.Log("should detect registry configuration rollback to default configuration")
		g.Expect(resourceClient.Delete(context.TODO(), &customDockerRegistryConfiguration)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Duration(0)))

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))

		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonJobsDeleted))

		assertSuccessfulFunctionBuild(t, resourceClient, reconciler, request, fnLabels, true)

		assertSuccessfulFunctionDeployment(t, resourceClient, reconciler, request, fnLabels, "registry.kyma.local", true)
	})
	t.Run("should set proper status on deployment fail", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		inFunction := newFixFunction(testNamespace, "deployment-fail", 1, 2)
		g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())
		defer deleteFunction(g, resourceClient, inFunction)

		fnLabels := reconciler.functionLabels(inFunction)
		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: inFunction.GetNamespace(), Name: inFunction.GetName()}}

		//WHEN
		t.Log("creating cm")
		_, err := reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating job")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		function := &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())

		jobList := &batchv1.JobList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(jobList.Items).To(gomega.HaveLen(1))

		job := &batchv1.Job{}
		g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
		g.Expect(job).ToNot(gomega.BeNil())
		job.Status.Succeeded = 1
		now := metav1.Now()
		job.Status.CompletionTime = &now
		g.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

		t.Log("job finished")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating deployment")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating svc")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating hpa")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("deployment failed")
		deployments := &appsv1.DeploymentList{}
		g.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
		g.Expect(len(deployments.Items)).To(gomega.Equal(1))
		deployment := &deployments.Items[0]
		deployment.Status.Conditions = []appsv1.DeploymentCondition{
			{Type: appsv1.DeploymentReplicaFailure, Status: corev1.ConditionTrue, Message: "Some random message", Reason: "some reason"},
			{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionFalse, Message: "Deployment doesn't have minimum availability."},
		}

		g.Expect(resourceClient.Status().Update(context.TODO(), deployment)).To(gomega.Succeed())

		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionFalse))

		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentFailed))
	})

	t.Run("should properly handle apiserver lags, when two resources are created by accident", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		inFunction := newFixFunction(testNamespace, "apiserver-lags", 1, 2)
		g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())
		defer deleteFunction(g, resourceClient, inFunction)

		fnLabels := reconciler.functionLabels(inFunction)
		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: inFunction.GetNamespace(), Name: inFunction.GetName()}}

		//WHEN
		t.Log("creating cm")
		_, err := reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		function := &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())

		configMapList := &corev1.ConfigMapList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, configMapList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(configMapList.Items).To(gomega.HaveLen(1))

		cm := configMapList.Items[0].DeepCopy()
		cm.Name = "" // generateName will create this
		cm.ResourceVersion = ""
		cm.UID = ""
		cm.CreationTimestamp = metav1.Time{}
		g.Expect(resourceClient.Create(context.TODO(), cm)).To(gomega.Succeed())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, configMapList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(configMapList.Items).To(gomega.HaveLen(2))

		t.Log("deleting all configMaps")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, configMapList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(configMapList.Items).To(gomega.HaveLen(0))

		t.Log("creating configMap again")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, configMapList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(configMapList.Items).To(gomega.HaveLen(1))

		t.Log("creating job")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())

		jobList := &batchv1.JobList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(jobList.Items).To(gomega.HaveLen(1))

		t.Log("accidently creating second job")
		excessJob := jobList.Items[0].DeepCopy()
		excessJob.Name = "" // generateName will create this
		excessJob.ResourceVersion = ""
		excessJob.UID = ""
		excessJob.CreationTimestamp = metav1.Time{}
		excessJob.Spec.Selector = nil
		delete(excessJob.Spec.Template.ObjectMeta.Labels, "controller-uid")
		delete(excessJob.Spec.Template.ObjectMeta.Labels, "job-name")
		g.Expect(resourceClient.Create(context.TODO(), excessJob)).To(gomega.Succeed())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(jobList.Items).To(gomega.HaveLen(2))

		t.Log("deleting all jobs")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(jobList.Items).To(gomega.HaveLen(0))

		t.Log("creating job again")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(jobList.Items).To(gomega.HaveLen(1))
		job := &batchv1.Job{}
		g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
		g.Expect(job).ToNot(gomega.BeNil())
		job.Status.Succeeded = 1
		now := metav1.Now()
		job.Status.CompletionTime = &now
		g.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

		t.Log("job finished")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating deployment")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		deployList := &appsv1.DeploymentList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, deployList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(deployList.Items).To(gomega.HaveLen(1))

		t.Log("creating next deployment by accident")

		excessDeploy := deployList.Items[0].DeepCopy()
		excessDeploy.Name = "" // generateName will create this
		excessDeploy.ResourceVersion = ""
		excessDeploy.UID = ""
		excessDeploy.CreationTimestamp = metav1.Time{}
		g.Expect(resourceClient.Create(context.TODO(), excessDeploy)).To(gomega.Succeed())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, deployList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(deployList.Items).To(gomega.HaveLen(2))

		t.Log("deleting excess deployment")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, deployList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(deployList.Items).To(gomega.HaveLen(0))

		t.Log("creating new deployment")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, deployList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(deployList.Items).To(gomega.HaveLen(1))

		t.Log("creating svc")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		svcList := &corev1.ServiceList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, svcList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(svcList.Items).To(gomega.HaveLen(1))

		t.Log("somehow there's been created a new svc with labels we use")
		excessSvc := svcList.Items[0].DeepCopy()
		excessSvc.Name = fmt.Sprintf("%s-%s", excessSvc.Name, "2")
		excessSvc.Spec.ClusterIP = ""
		excessSvc.ResourceVersion = ""
		excessSvc.UID = ""
		excessSvc.CreationTimestamp = metav1.Time{}
		g.Expect(resourceClient.Create(context.TODO(), excessSvc)).To(gomega.Succeed())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, svcList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(svcList.Items).To(gomega.HaveLen(2))

		t.Log("deleting that svc")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, svcList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(svcList.Items).To(gomega.HaveLen(1))

		t.Log("creating hpa")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		hpaList := &autoscalingv1.HorizontalPodAutoscalerList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(deployList.Items).To(gomega.HaveLen(1))
		t.Log("creating next hpa by accident - apiserver lag")

		g.Expect(hpaList.Items).To(gomega.HaveLen(1))
		excessHpa := hpaList.Items[0].DeepCopy()
		excessHpa.Name = "" // generateName will create this
		excessHpa.ResourceVersion = ""
		excessHpa.UID = ""
		excessHpa.CreationTimestamp = metav1.Time{}
		g.Expect(resourceClient.Create(context.TODO(), excessHpa)).To(gomega.Succeed())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(hpaList.Items).To(gomega.HaveLen(2))

		t.Log("deleting excess hpa 🔫")

		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(hpaList.Items).To(gomega.HaveLen(0))

		t.Log("creating new hpa")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(hpaList.Items).To(gomega.HaveLen(1))

		t.Log("deployment ready")
		deployments := &appsv1.DeploymentList{}
		g.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
		g.Expect(len(deployments.Items)).To(gomega.Equal(1))
		deployment := &deployments.Items[0]

		deployment.Status.Conditions = []appsv1.DeploymentCondition{
			{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: MinimumReplicasAvailable},
			{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: NewRSAvailableReason},
		}
		g.Expect(resourceClient.Status().Update(context.TODO(), deployment)).To(gomega.Succeed())

		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		g.Expect(hpaList.Items[0].Spec.ScaleTargetRef.Name).To(gomega.Equal(deployList.Items[0].Name), "hpa should target proper deployment")

		t.Log("deleting deployment by 'accident' to check proper hpa-deployment reference")

		err = resourceClient.DeleteAllBySelector(context.TODO(), &appsv1.Deployment{}, request.Namespace, labels.SelectorFromSet(fnLabels))
		g.Expect(err).To(gomega.BeNil())

		t.Log("recreating deployment")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("updating hpa")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(hpaList.Items).To(gomega.HaveLen(1))

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, deployList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(deployList.Items).To(gomega.HaveLen(1))

		g.Expect(hpaList.Items[0].Spec.ScaleTargetRef.Name).To(gomega.Equal(deployList.Items[0].Name), "hpa should target proper deployment")
	})

	t.Run("should requeue before creating a job", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		inFunction := newFixFunction(testNamespace, "requeue-before-job", 1, 2)
		g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())
		defer deleteFunction(g, resourceClient, inFunction)

		// Create new reconciler as this test modify reconciler configuration MaxSimultaneousJobs value
		statsCollector := &automock.StatsCollector{}
		statsCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()

		reconciler := NewFunction(resourceClient, log.Log, testCfg, nil, record.NewFakeRecorder(100), statsCollector, make(chan bool))
		reconciler.config.Build.MaxSimultaneousJobs = 1

		fnLabels := reconciler.functionLabels(inFunction)

		secondFunction := newFixFunction(testNamespace, "second-function", 1, 2)
		secondRequest := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: secondFunction.GetNamespace(), Name: secondFunction.GetName()}}
		g.Expect(resourceClient.Create(context.TODO(), secondFunction)).To(gomega.Succeed())

		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: inFunction.GetNamespace(), Name: inFunction.GetName()}}

		//WHEN
		t.Log("creating 2 cms")
		_, err := reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		_, err = reconciler.Reconcile(ctx, secondRequest)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating 2 jobs")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		result, err := reconciler.Reconcile(ctx, secondRequest)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.RequeueAfter).To(gomega.BeIdenticalTo(time.Second * 5))

		t.Log("handling first job")
		function := &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())

		jobList := &batchv1.JobList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(jobList.Items).To(gomega.HaveLen(1))

		job := &batchv1.Job{}
		g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
		g.Expect(job).ToNot(gomega.BeNil())
		job.Status.Succeeded = 1
		now := metav1.Now()
		job.Status.CompletionTime = &now
		g.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

		t.Log("first job finished")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("handling second job")
		secFunction := &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, secFunction)).To(gomega.Succeed())

		_, err = reconciler.Reconcile(ctx, secondRequest)
		g.Expect(err).To(gomega.BeNil())

		secJobList := &batchv1.JobList{}
		err = reconciler.client.ListByLabel(context.TODO(), secFunction.GetNamespace(), reconciler.internalFunctionLabels(secondFunction), secJobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(secJobList.Items).To(gomega.HaveLen(1))

		secJob := &batchv1.Job{}
		g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, secJob)).To(gomega.Succeed())
		g.Expect(secJob).ToNot(gomega.BeNil())
		secJob.Status.Succeeded = 1
		now = metav1.Now()
		secJob.Status.CompletionTime = &now
		g.Expect(resourceClient.Status().Update(context.TODO(), secJob)).To(gomega.Succeed())

		t.Log("second job finished")
		_, err = reconciler.Reconcile(ctx, secondRequest)
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("should behave correctly on label addition and subtraction", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		inFunction := newFixFunction(testNamespace, "labels-operations", 1, 2)
		g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())
		defer deleteFunction(g, resourceClient, inFunction)

		fnLabels := reconciler.functionLabels(inFunction)
		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: inFunction.GetNamespace(), Name: inFunction.GetName()}}

		//WHEN
		t.Log("creating cm")
		_, err := reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		function := &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())

		t.Log("creating job")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		jobList := &batchv1.JobList{}
		err = resourceClient.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(jobList.Items).To(gomega.HaveLen(1))

		job := &batchv1.Job{}
		g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
		g.Expect(job).ToNot(gomega.BeNil())
		job.Status.Succeeded = 1
		now := metav1.Now()
		job.Status.CompletionTime = &now
		g.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

		t.Log("job finished")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating deployment")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating svc")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating hpa")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("deployment ready")
		deployments := &appsv1.DeploymentList{}
		g.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
		g.Expect(len(deployments.Items)).To(gomega.Equal(1))

		deployment := &deployments.Items[0]
		deployment.Status.Conditions = []appsv1.DeploymentCondition{
			{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: MinimumReplicasAvailable},
			{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: NewRSAvailableReason},
		}
		g.Expect(resourceClient.Status().Update(context.TODO(), deployment)).To(gomega.Succeed())

		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("updating function metadata.labels")
		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Labels).To(gomega.BeNil())

		functionWithLabels := function.DeepCopy()
		functionWithLabels.Labels = map[string]string{
			addedLabelKey: addedLabelValue,
		}

		g.Expect(resourceClient.Update(context.TODO(), functionWithLabels)).To(gomega.Succeed())

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Labels).NotTo(gomega.BeNil())

		t.Log("updating configmap")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		configMapList := &corev1.ConfigMapList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, configMapList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(configMapList.Items).To(gomega.HaveLen(1))
		g.Expect(configMapList.Items[0].Labels).To(gomega.HaveLen(4))

		cmLabelVal, ok := configMapList.Items[0].Labels[addedLabelKey]
		g.Expect(ok).To(gomega.BeTrue())
		g.Expect(cmLabelVal).To(gomega.Equal(addedLabelValue))

		t.Log("updating job")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonJobUpdated))

		jobList = &batchv1.JobList{}
		err = resourceClient.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(jobList.Items).To(gomega.HaveLen(1))
		g.Expect(jobList.Items[0].Labels).To(gomega.HaveLen(4))

		jobLabelVal, ok := jobList.Items[0].Labels[addedLabelKey]
		g.Expect(ok).To(gomega.BeTrue())
		g.Expect(jobLabelVal).To(gomega.Equal(addedLabelValue))

		t.Log("reconciling job to make sure it's already finished")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonJobFinished))

		t.Log("updating deployment")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		deployList := &appsv1.DeploymentList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, deployList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(deployList.Items).To(gomega.HaveLen(1))
		g.Expect(deployList.Items[0].Labels).To(gomega.HaveLen(4))

		deployLabelVal, ok := deployList.Items[0].Labels[addedLabelKey]
		g.Expect(ok).To(gomega.BeTrue())
		g.Expect(deployLabelVal).To(gomega.Equal(addedLabelValue))

		t.Log("updating service")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		svcList := &corev1.ServiceList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, svcList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(svcList.Items).To(gomega.HaveLen(1))
		g.Expect(svcList.Items[0].Labels).To(gomega.HaveLen(4))

		svcLabelVal, ok := svcList.Items[0].Labels[addedLabelKey]
		g.Expect(ok).To(gomega.BeTrue())
		g.Expect(svcLabelVal).To(gomega.Equal(addedLabelValue))

		t.Log("updating hpa")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		hpaList := &autoscalingv1.HorizontalPodAutoscalerList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(hpaList.Items).To(gomega.HaveLen(1))
		g.Expect(hpaList.Items[0].Labels).To(gomega.HaveLen(4))

		hpaLabelVal, ok := hpaList.Items[0].Labels[addedLabelKey]
		g.Expect(ok).To(gomega.BeTrue())
		g.Expect(hpaLabelVal).To(gomega.Equal(addedLabelValue))

		t.Log("status ready")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))

		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentReady))

		t.Log("getting rid of formerly added labels")
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Labels).NotTo(gomega.BeNil())

		functionWithoutLabels := function.DeepCopy()
		functionWithoutLabels.Labels = nil
		g.Expect(resourceClient.Update(context.TODO(), functionWithoutLabels)).To(gomega.Succeed())

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Labels).To(gomega.BeNil())

		t.Log("reconciling again -> configmap")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		configMapList = &corev1.ConfigMapList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, configMapList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(configMapList.Items).To(gomega.HaveLen(1))

		_, ok = configMapList.Items[0].Labels[addedLabelKey]
		g.Expect(ok).To(gomega.BeFalse())

		t.Log("reconciling again -> job")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		jobList = &batchv1.JobList{}
		err = resourceClient.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(jobList.Items).To(gomega.HaveLen(1))

		_, ok = jobList.Items[0].Labels[addedLabelKey]
		g.Expect(ok).To(gomega.BeFalse())

		t.Log("reconciling again -> job finished")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("reconciling again -> deployment")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		deployList = &appsv1.DeploymentList{}
		err = resourceClient.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, deployList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(deployList.Items).To(gomega.HaveLen(1))

		_, ok = deployList.Items[0].Labels[addedLabelKey]
		g.Expect(ok).To(gomega.BeFalse())

		t.Log("reconciling again -> service")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		svcList = &corev1.ServiceList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, svcList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(svcList.Items).To(gomega.HaveLen(1))

		_, ok = svcList.Items[0].Labels[addedLabelKey]
		g.Expect(ok).To(gomega.BeFalse())

		t.Log("reconciling again -> hpa")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		hpaList = &autoscalingv1.HorizontalPodAutoscalerList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(hpaList.Items).To(gomega.HaveLen(1))

		_, ok = hpaList.Items[0].Labels[addedLabelKey]
		g.Expect(ok).To(gomega.BeFalse())

		t.Log("reconciling again -> deployment ready")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))

		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentReady))
	})

	t.Run("should behave correctly on label addition when job is in building phase", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		inFunction := newFixFunction(testNamespace, "add-label-while-building", 1, 2)
		g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())
		defer deleteFunction(g, resourceClient, inFunction)

		fnLabels := reconciler.functionLabels(inFunction)
		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: inFunction.GetNamespace(), Name: inFunction.GetName()}}

		//WHEN
		t.Log("creating cm")
		_, err := reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating job")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		function := &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())

		jobList := &batchv1.JobList{}
		err = resourceClient.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(jobList.Items).To(gomega.HaveLen(1))

		t.Log("updating function metadata.labels")
		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Labels).To(gomega.BeNil())

		functionWithLabels := function.DeepCopy()
		functionWithLabels.Labels = map[string]string{
			"that-label": "wasnt-here",
		}
		g.Expect(resourceClient.Update(context.TODO(), functionWithLabels)).To(gomega.Succeed())

		function = &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Labels).NotTo(gomega.BeNil())

		t.Log("updating configmap")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("updating job")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("job's finished")

		job := &batchv1.Job{}
		g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
		g.Expect(job).ToNot(gomega.BeNil())
		job.Status.Succeeded = 1
		now := metav1.Now()
		job.Status.CompletionTime = &now
		g.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating deployment")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating svc")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("creating hpa")
		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("deployment ready")
		deployments := &appsv1.DeploymentList{}
		g.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
		g.Expect(len(deployments.Items)).To(gomega.Equal(1))

		deployment := &deployments.Items[0]
		deployment.Status.Conditions = []appsv1.DeploymentCondition{
			{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: MinimumReplicasAvailable},
			{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: NewRSAvailableReason},
		}
		g.Expect(resourceClient.Status().Update(context.TODO(), deployment)).To(gomega.Succeed())

		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("should handle reconcilation lags", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)

		//WHEN
		t.Log("handling not existing Function")
		result, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "nope", Name: "noooooopppeee"}})

		//THEN
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))
	})

	t.Run("should return error when desired dockerfile runtime configmap not found", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		testNamespace := "test-namespace"
		fnName := "function"
		function := newFixFunction(testNamespace, fnName, 1, 2)
		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: function.GetNamespace(), Name: function.GetName()}}
		g.Expect(resourceClient.Create(context.TODO(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		}})).To(gomega.Succeed())
		g.Expect(resourceClient.Create(context.TODO(), function)).To(gomega.Succeed())
		defer deleteFunction(g, resourceClient, function)

		//WHEN
		_, err := reconciler.Reconcile(ctx, request)

		//THEN
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(err.Error()).To(gomega.ContainSubstring("Expected one config map, found 0"))

	})

	t.Run("should use HPA only when needed", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		inFunction := newFixFunction(testNamespace, "hpa", 1, 2)
		g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())
		defer deleteFunction(g, resourceClient, inFunction)

		fnLabels := reconciler.functionLabels(inFunction)
		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: inFunction.GetNamespace(), Name: inFunction.GetName()}}

		//WHEN
		t.Log("creating the ConfigMap")
		result, err := reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function := &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(1))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonConfigMapCreated))

		configMapList := &corev1.ConfigMapList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, configMapList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(configMapList.Items).To(gomega.HaveLen(1))
		g.Expect(configMapList.Items[0].Data[FunctionSourceKey]).To(gomega.Equal(function.Spec.Source))
		g.Expect(configMapList.Items[0].Data[FunctionDepsKey]).To(gomega.Equal("{}"))

		assertSuccessfulFunctionBuild(t, resourceClient, reconciler, request, fnLabels, false)

		assertSuccessfulFunctionDeployment(t, resourceClient, reconciler, request, fnLabels, "registry.kyma.local", false)

		two := int32(2)
		four := int32(4)

		t.Log("updating function to use fixed replicas number")
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		function.Spec.MinReplicas = &two
		function.Spec.MaxReplicas = &two
		g.Expect(resourceClient.Update(context.TODO(), function)).To(gomega.Succeed())

		t.Log("updating deployment with new number of replicas")
		result, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))
		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentUpdated))
		deployments := &appsv1.DeploymentList{}
		g.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
		g.Expect(len(deployments.Items)).To(gomega.Equal(1))
		deployment := &deployments.Items[0]
		g.Expect(deployment).ToNot(gomega.BeNil())
		g.Expect(deployment.Spec.Replicas).To(gomega.Equal(&two))

		t.Log("removing hpa")
		result, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		hpaList := &autoscalingv1.HorizontalPodAutoscalerList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(hpaList.Items).To(gomega.HaveLen(0))

		t.Log("deployment ready")
		result, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Minute * 5))

		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentReady))

		t.Log("updating function to use scalable replicas number")
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		function.Spec.MaxReplicas = &four
		g.Expect(resourceClient.Update(context.TODO(), function)).To(gomega.Succeed())

		t.Log("creating hpa")
		result, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(time.Duration(0)))

		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonHorizontalPodAutoscalerCreated))

		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(hpaList.Items).To(gomega.HaveLen(1))

		hpaSpec := hpaList.Items[0].Spec

		g.Expect(hpaSpec.ScaleTargetRef.Name).To(gomega.Equal(deployment.GetName()))
		g.Expect(hpaSpec.ScaleTargetRef.Kind).To(gomega.Equal("Deployment"))
		g.Expect(hpaSpec.ScaleTargetRef.APIVersion).To(gomega.Equal(appsv1.SchemeGroupVersion.String()))

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

		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentReady))
	})
	t.Run("should properly handle `kubectl rollout restart` changing annotations in deployment", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		inFunction := newFixFunction(testNamespace, "rollout-restart-fn", 1, 2)
		g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())
		defer deleteFunction(g, resourceClient, inFunction)

		fnLabels := reconciler.functionLabels(inFunction)
		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: inFunction.GetNamespace(), Name: inFunction.GetName()}}

		//WHEN
		t.Log("succesfully deploying a function")
		_, err := reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())
		assertSuccessfulFunctionBuild(t, resourceClient, reconciler, request, fnLabels, false)
		assertSuccessfulFunctionDeployment(t, resourceClient, reconciler, request, fnLabels, "registry.kyma.local", false)

		t.Log("updating deployment.spec.template.metadata.annotations, e.g. by using kubectl rollout restart command")
		deployments := &appsv1.DeploymentList{}
		g.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
		g.Expect(len(deployments.Items)).To(gomega.Equal(1))
		deployment := &deployments.Items[0]
		g.Expect(deployment).ToNot(gomega.BeNil())

		g.Expect(deployment.Spec.Template.Annotations).To(gomega.Equal(map[string]string{
			"proxy.istio.io/config": "{ \"holdApplicationUntilProxyStarts\": true }",
		}))
		copiedDeploy := deployment.DeepCopy()
		restartedAtAnnotation := map[string]string{
			"kubectl.kubernetes.io/restartedAt": "2021-03-10T11:28:01+01:00", // example annotation added by kubectl
		}
		copiedDeploy.Spec.Template.Annotations = restartedAtAnnotation
		g.Expect(resourceClient.Update(context.Background(), copiedDeploy))

		_, err = reconciler.Reconcile(ctx, request)
		g.Expect(err).To(gomega.BeNil())

		t.Log("making sure function is ready")
		function := &serverlessv1alpha1.Function{}
		g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		g.Expect(function.Status.Conditions).To(gomega.HaveLen(conditionLen))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))
		g.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonDeploymentReady))

		t.Log("checking whether that added annotation is still there")
		deployments = &appsv1.DeploymentList{}
		g.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
		g.Expect(len(deployments.Items)).To(gomega.Equal(1))
		deployment = &deployments.Items[0]
		g.Expect(deployment).ToNot(gomega.BeNil())

		g.Expect(deployment.Spec.Template.Annotations).To(gomega.Equal(restartedAtAnnotation))
	})
}

func deleteFunction(g *gomega.GomegaWithT, resourceClient resource.Client, function *serverlessv1alpha1.Function) {
	err := resourceClient.Delete(context.TODO(), function)
	g.Expect(err).To(gomega.BeNil())
}
