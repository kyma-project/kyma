package serverless

import (
	"context"
	"fmt"
	"math/rand"
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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/automock"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var _ = ginkgo.Describe("Function", func() {
	var newFixFunction = func(namespace, name string, minReplicas, maxReplicas int) *serverlessv1alpha1.Function {
		one := int32(minReplicas)
		two := int32(maxReplicas)
		suffix := rand.Int()

		return &serverlessv1alpha1.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", name, suffix),
				Namespace: namespace,
			},
			Spec: serverlessv1alpha1.FunctionSpec{
				SourceType: serverlessv1alpha1.SourceTypeGit,
				Source:     fmt.Sprintf("%s-%d", name, suffix),
				Repository: serverlessv1alpha1.Repository{
					BaseDir:   "/",
					Runtime:   serverlessv1alpha1.RuntimeNodeJS12,
					Reference: "master",
				},
				Env: []corev1.EnvVar{
					{
						Name:  "TEST_1",
						Value: "VAL_1",
					},
					{
						Name:  "TEST_2",
						Value: "VAL_2",
					},
				},
				Resources:   corev1.ResourceRequirements{},
				MinReplicas: &one,
				MaxReplicas: &two,
				Labels: map[string]string{
					testBindingLabel1: "foobar",
					testBindingLabel2: testBindingLabelValue,
					"foo":             "bar",
				},
			},
		}
	}

	var newFixRepository = func(name, namespace string) *serverlessv1alpha1.GitRepository {
		return &serverlessv1alpha1.GitRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: serverlessv1alpha1.GitRepositorySpec{
				URL: "https://mock.repo/kyma/test",
				Auth: &serverlessv1alpha1.RepositoryAuth{
					Type:       serverlessv1alpha1.RepositoryAuthBasic,
					SecretName: name,
				},
			},
		}
	}

	var newFixSecret = func(name, namespace string) *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			StringData: map[string]string{
				"user":     "test",
				"password": "test",
			},
		}
	}

	var fixMockGitOperator = func(secretName string) *automock.GitOperator {
		options := git.Options{
			URL:       "https://mock.repo/kyma/test",
			Reference: "master",
			Auth: &git.AuthOptions{
				Type: git.RepositoryAuthBasic,
				Credentials: map[string]string{
					"user":     "test",
					"password": "test",
				},
				SecretName: secretName,
			},
		}

		mock := new(automock.GitOperator)
		mock.On("LastCommit", options).Return("pierwszy-hash", nil)
		options.Reference = "newone"
		mock.On("LastCommit", options).Return("a376218bdcd705cc39aa7ce7f310769fab6d51c9", nil)

		return mock
	}

	var (
		reconciler *FunctionReconciler
		request    ctrl.Request
		fnLabels   map[string]string
	)

	ginkgo.BeforeEach(func() {
		function := newFixFunction("tutaj-devops", "ah-tak-przeciez", 1, 2)
		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: function.GetNamespace(), Name: function.GetName()}}
		gomega.Expect(resourceClient.Create(context.TODO(), function)).To(gomega.Succeed())

		reconciler = NewFunction(resourceClient, log.Log, config, record.NewFakeRecorder(100))
		reconciler.gitOperator = fixMockGitOperator(function.Name)
		fnLabels = reconciler.internalFunctionLabels(function)

		repo := newFixRepository(function.GetName(), "tutaj-devops")
		gomega.Expect(resourceClient.Create(context.TODO(), repo)).To(gomega.Succeed())

		secret := newFixSecret(function.Name, "tutaj-devops")
		gomega.Expect(resourceClient.Create(context.TODO(), secret)).To(gomega.Succeed())
	})

	ginkgo.It("should handle reconcilation lags", func() {
		ginkgo.By("handling not existing Function")
		result, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "nope", Name: "noooooopppeee"}})
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))
	})

	ginkgo.It("should successfully update Function", func() {
		ginkgo.By("creating the Function")
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

		gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonSourceUpdated))

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

		ginkgo.By("change function branch")
		function.Spec.Reference = "newone"
		gomega.Expect(resourceClient.Update(context.TODO(), function)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		// check if status was updated
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(2))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(function.Status.Reference).To(gomega.Equal("newone"))
		gomega.Expect(function.Status.Commit).To(gomega.Equal("a376218bdcd705cc39aa7ce7f310769fab6d51c9"))

		ginkgo.By("delete the old Job")
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(2))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))

		jobList = &batchv1.JobList{}
		err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(jobList.Items).To(gomega.HaveLen(0))

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

		jobList = &batchv1.JobList{}
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
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(2))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		gomega.Expect(reconciler.getConditionReason(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(serverlessv1alpha1.ConditionReasonJobRunning))

		ginkgo.By("build finished")
		job = &batchv1.Job{}
		gomega.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
		gomega.Expect(job).ToNot(gomega.BeNil())
		job.Status.Succeeded = 1
		now = metav1.Now()
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
		gomega.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
		gomega.Expect(len(deployments.Items)).To(gomega.Equal(1))
		deployment := &deployments.Items[0]
		gomega.Expect(deployment).ToNot(gomega.BeNil())
		gomega.Expect(deployment.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
		gomega.Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(gomega.Equal(reconciler.buildImageAddress(function)))
		gomega.Expect(deployment.Spec.Template.Labels).To(gomega.HaveLen(7))
		gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionNameLabel]).To(gomega.Equal(function.Name))
		gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionManagedByLabel]).To(gomega.Equal("function-controller"))
		gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionUUIDLabel]).To(gomega.Equal(string(function.UID)))
		gomega.Expect(deployment.Spec.Template.Labels[serverlessv1alpha1.FunctionResourceLabel]).To(gomega.Equal(serverlessv1alpha1.FunctionResourceLabelDeploymentValue))
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

		gomega.Expect(labels.AreLabelsInWhiteList(svc.Spec.Selector, job.Spec.Template.Labels)).To(gomega.BeFalse(), "svc selector should not catch job pods")
		gomega.Expect(svc.Spec.Selector).To(gomega.Equal(deployment.Spec.Selector.MatchLabels))

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
})
