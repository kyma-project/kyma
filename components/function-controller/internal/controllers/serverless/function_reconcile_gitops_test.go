package serverless

import (
	"context"
	"fmt"
	"log"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/mock"

	git2go "github.com/libgit2/git2go/v31"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/automock"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
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
)

type testDataScenario struct {
	info       string
	authType   *string
	stringData map[string]string
}

var (
	authTypeBasic = "basic"
	authTypeKey   = "key"
)

var testDataScenarios = []testDataScenario{
	{
		info:     "auth-key-with-pw",
		authType: &authTypeKey,
		stringData: map[string]string{
			"user":     "test",
			"password": "test",
		},
	},
	{
		info:     "auth-basic",
		authType: &authTypeBasic,
		stringData: map[string]string{
			"user":     "test",
			"password": "test",
		},
	},
	{
		info:     "auth-key",
		authType: &authTypeKey,
		stringData: map[string]string{
			"authTypeKey": "123",
		},
	},
	{
		info:       "no-auth",
		authType:   nil,
		stringData: nil,
	},
}

var newMockedGitClient = func(auth *git.AuthOptions) *automock.GitClient {
	options := git.Options{
		URL:       "https://mock.repo/kyma/test",
		Reference: "main",
	}

	options.Auth = auth
	log.Println(options)
	m := new(automock.GitClient)

	m.On("LastCommit", options).Return("pierwszy-hash", nil)
	options.Reference = "newone"
	m.On("LastCommit", options).Return("a376218bdcd705cc39aa7ce7f310769fab6d51c9", nil)

	return m
}

func TestGitOpsWithContinuousGitCheckout(t *testing.T) {
	//GIVEN
	continuousGitCheckout := true

	g := gomega.NewGomegaWithT(t)
	rtm := serverlessv1alpha2.NodeJs18
	resourceClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	testCfg := setUpControllerConfig(g)
	initializeServerlessResources(g, resourceClient)
	createDockerfileForRuntime(g, resourceClient, rtm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, testData := range testDataScenarios {
		t.Run(fmt.Sprintf("[%s] should successfully update Function]", testData.info), func(t *testing.T) {
			//GIVEN
			g := gomega.NewGomegaWithT(t)
			name := fmt.Sprintf("test-me-plz-%d", i)

			var auth *serverlessv1alpha2.RepositoryAuth
			if testData.authType != nil {
				auth = &serverlessv1alpha2.RepositoryAuth{
					Type:       serverlessv1alpha2.RepositoryAuthType(*testData.authType),
					SecretName: name,
				}
				secret := newTestSecret(name, testNamespace, testData.stringData)
				g.Expect(resourceClient.Create(context.TODO(), secret)).To(gomega.Succeed())

			}

			inFunction := newTestGitFunction(testNamespace, name, auth, 1, 2, continuousGitCheckout)
			g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())

			request := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: inFunction.GetNamespace(),
					Name:      inFunction.GetName(),
				},
			}

			var gitAuthOpts *git.AuthOptions
			if testData.authType != nil {
				gitAuthOpts = &git.AuthOptions{
					Type:        git.RepositoryAuthType(*testData.authType),
					Credentials: testData.stringData,
					SecretName:  name,
				}
			}

			gitClient := newMockedGitClient(gitAuthOpts)
			factory := automock.NewGitClientFactory(t)
			factory.On("GetGitClient", mock.Anything).Return(gitClient)
			defer factory.AssertExpectations(t)

			statsCollector := &automock.StatsCollector{}
			statsCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()

			reconciler := &FunctionReconciler{
				Log:               zap.NewNop().Sugar(),
				client:            resourceClient,
				recorder:          record.NewFakeRecorder(100),
				config:            testCfg,
				gitFactory:        factory,
				statsCollector:    statsCollector,
				initStateFunction: stateFnGitCheckSources,
			}

			fnLabels := reconciler.internalFunctionLabels(inFunction)

			//WHEN
			t.Log("creating the Function")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)
			// verify function
			function := &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(1))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonSourceUpdated)

			t.Log("creating the Job")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)

			jobList := &batchv1.JobList{}
			err := reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
			g.Expect(err).To(gomega.BeNil())
			g.Expect(jobList.Items).To(gomega.HaveLen(1))

			t.Log("build in progress")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonJobRunning)

			t.Log("build finished")
			job := &batchv1.Job{}
			g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{
				Namespace: jobList.Items[0].GetNamespace(),
				Name:      jobList.Items[0].GetName(),
			}, job)).To(gomega.Succeed())

			g.Expect(job).ToNot(gomega.BeNil())
			job.Status.Succeeded = 1
			now := metav1.Now()
			job.Status.CompletionTime = &now
			g.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonJobFinished)

			t.Log("change function branch")
			function.Spec.Source.GitRepository.Reference = "newone"
			g.Expect(resourceClient.Update(context.TODO(), function)).To(gomega.Succeed())

			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			// check if status was updated
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveStatusReference("newone"))
			g.Expect(function).To(haveStatusCommit("a376218bdcd705cc39aa7ce7f310769fab6d51c9"))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)

			t.Log("delete the old Job")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionRunning)

			jobList = &batchv1.JobList{}
			err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
			g.Expect(err).To(gomega.BeNil())
			g.Expect(jobList.Items).To(gomega.HaveLen(0))

			t.Log("creating the Job")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)

			jobList = &batchv1.JobList{}
			err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
			g.Expect(err).To(gomega.BeNil())
			g.Expect(jobList.Items).To(gomega.HaveLen(1))

			t.Log("build in progress")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonJobRunning)

			t.Log("build finished")
			job = &batchv1.Job{}
			g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{
				Namespace: jobList.Items[0].GetNamespace(),
				Name:      jobList.Items[0].GetName(),
			}, job)).To(gomega.Succeed())
			g.Expect(job).ToNot(gomega.BeNil())
			job.Status.Succeeded = 1
			now = metav1.Now()
			job.Status.CompletionTime = &now
			g.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonJobFinished)

			t.Log("deploy started")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)

			deployments := &appsv1.DeploymentList{}
			g.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).
				To(gomega.Succeed())

			g.Expect(len(deployments.Items)).To(gomega.Equal(1))

			deployment := &deployments.Items[0]

			s := systemState{
				//TODO https://github.com/kyma-project/kyma/issues/14079
				instance: *function,
			}

			expectedImage := s.buildImageAddress("localhost:32132")
			g.Expect(deployment).To(gomega.Not(gomega.BeNil()))
			g.Expect(deployment).To(haveSpecificContainer0Image(expectedImage))
			g.Expect(deployment).To(haveLabelLen(7))
			g.Expect(deployment).To(haveLabelWithValue(serverlessv1alpha2.FunctionNameLabel, function.Name))
			g.Expect(deployment).To(haveLabelWithValue(serverlessv1alpha2.FunctionManagedByLabel, serverlessv1alpha2.FunctionControllerValue))
			g.Expect(deployment).To(haveLabelWithValue(serverlessv1alpha2.FunctionUUIDLabel, string(function.UID)))
			g.Expect(deployment).To(haveLabelWithValue(
				serverlessv1alpha2.FunctionResourceLabel, serverlessv1alpha2.FunctionResourceLabelDeploymentValue))

			g.Expect(deployment).To(haveLabelWithValue(testBindingLabel1, "foobar"))
			g.Expect(deployment).To(haveLabelWithValue(testBindingLabel2, testBindingLabelValue))
			g.Expect(deployment).To(haveLabelWithValue("foo", "bar"))

			t.Log("service creation")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonServiceCreated)

			svc := &corev1.Service{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, svc)).To(gomega.Succeed())
			g.Expect(err).To(gomega.BeNil())

			g.Expect(svc.Spec.Ports).To(gomega.HaveLen(1))
			g.Expect(svc.Spec.Ports[0].Name).To(gomega.Equal("http"))
			g.Expect(svc.Spec.Ports[0].TargetPort).To(gomega.Equal(intstr.FromInt(8080)))

			g.Expect(isSubset(svc.Spec.Selector, job.Spec.Template.Labels)).
				To(gomega.BeFalse(), "svc selector should not catch job pods")

			g.Expect(svc.Spec.Selector).To(gomega.Equal(deployment.Spec.Selector.MatchLabels))

			t.Log("HPA creation")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonHorizontalPodAutoscalerCreated)

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

			g.Expect(reconciler.Reconcile(ctx, request)).To(beFinishedReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveConditionRunning)
			g.Expect(function).To(haveConditionReasonDeploymentReady)

			t.Log("should not change state on reconcile")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beFinishedReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveConditionRunning)
			g.Expect(function).To(haveConditionReasonDeploymentReady)
		})
	}
}

func TestGitOpsWithoutContinuousGitCheckout(t *testing.T) {
	//GIVEN
	continuousGitCheckout := false

	g := gomega.NewGomegaWithT(t)
	rtm := serverlessv1alpha2.NodeJs18
	resourceClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	testCfg := setUpControllerConfig(g)
	initializeServerlessResources(g, resourceClient)
	createDockerfileForRuntime(g, resourceClient, rtm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, testData := range testDataScenarios {
		t.Run(fmt.Sprintf("[%s] should successfully update Function]", testData.info), func(t *testing.T) {
			//GIVEN
			g := gomega.NewGomegaWithT(t)
			name := fmt.Sprintf("test-me-plz-%d", i)

			var auth *serverlessv1alpha2.RepositoryAuth
			if testData.authType != nil {
				auth = &serverlessv1alpha2.RepositoryAuth{
					Type:       serverlessv1alpha2.RepositoryAuthType(*testData.authType),
					SecretName: name,
				}
				secret := newTestSecret(name, testNamespace, testData.stringData)
				g.Expect(resourceClient.Create(context.TODO(), secret)).To(gomega.Succeed())

			}

			inFunction := newTestGitFunction(testNamespace, name, auth, 1, 2, continuousGitCheckout)
			g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())

			request := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: inFunction.GetNamespace(),
					Name:      inFunction.GetName(),
				},
			}

			var gitAuthOpts *git.AuthOptions
			if testData.authType != nil {
				gitAuthOpts = &git.AuthOptions{
					Type:        git.RepositoryAuthType(*testData.authType),
					Credentials: testData.stringData,
					SecretName:  name,
				}
			}

			gitClient := newMockedGitClient(gitAuthOpts)
			factory := automock.NewGitClientFactory(t)
			factory.On("GetGitClient", mock.Anything).Return(gitClient)
			defer factory.AssertExpectations(t)

			statsCollector := &automock.StatsCollector{}
			statsCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()

			reconciler := &FunctionReconciler{
				Log:               zap.NewNop().Sugar(),
				client:            resourceClient,
				recorder:          record.NewFakeRecorder(100),
				config:            testCfg,
				gitFactory:        factory,
				statsCollector:    statsCollector,
				initStateFunction: stateFnGitCheckSources,
			}

			fnLabels := reconciler.internalFunctionLabels(inFunction)

			//WHEN
			t.Log("creating the Function")

			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)
			// verify function
			function := &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(1))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonSourceUpdated)

			t.Log("creating the Job")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)

			jobList := &batchv1.JobList{}
			err := reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
			g.Expect(err).To(gomega.BeNil())
			g.Expect(jobList.Items).To(gomega.HaveLen(1))

			t.Log("build in progress")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonJobRunning)

			t.Log("build finished")
			job := &batchv1.Job{}
			g.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{
				Namespace: jobList.Items[0].GetNamespace(),
				Name:      jobList.Items[0].GetName(),
			}, job)).To(gomega.Succeed())

			g.Expect(job).ToNot(gomega.BeNil())
			job.Status.Succeeded = 1
			now := metav1.Now()
			job.Status.CompletionTime = &now
			g.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonJobFinished)

			t.Log("Deployment is created")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)

			deployments := &appsv1.DeploymentList{}
			g.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).
				To(gomega.Succeed())

			g.Expect(len(deployments.Items)).To(gomega.Equal(1))

			deployment := &deployments.Items[0]

			s := systemState{
				//TODO https://github.com/kyma-project/kyma/issues/14079
				instance: *function,
			}

			expectedImage := s.buildImageAddress("localhost:32132")
			g.Expect(deployment).To(gomega.Not(gomega.BeNil()))
			g.Expect(deployment).To(haveSpecificContainer0Image(expectedImage))
			g.Expect(deployment).To(haveLabelLen(7))
			g.Expect(deployment).To(haveLabelWithValue(serverlessv1alpha2.FunctionNameLabel, function.Name))
			g.Expect(deployment).To(haveLabelWithValue(serverlessv1alpha2.FunctionManagedByLabel, serverlessv1alpha2.FunctionControllerValue))
			g.Expect(deployment).To(haveLabelWithValue(serverlessv1alpha2.FunctionUUIDLabel, string(function.UID)))
			g.Expect(deployment).To(haveLabelWithValue(
				serverlessv1alpha2.FunctionResourceLabel, serverlessv1alpha2.FunctionResourceLabelDeploymentValue))

			g.Expect(deployment).To(haveLabelWithValue(testBindingLabel1, "foobar"))
			g.Expect(deployment).To(haveLabelWithValue(testBindingLabel2, testBindingLabelValue))
			g.Expect(deployment).To(haveLabelWithValue("foo", "bar"))

			t.Log("change function branch")
			function.Spec.Source.GitRepository.Reference = "newone"
			g.Expect(resourceClient.Update(context.TODO(), function)).To(gomega.Succeed())

			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			// check if status was updated
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(3))
			g.Expect(function).To(haveStatusReference("main"))
			g.Expect(function).To(haveStatusCommit("pierwszy-hash"))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)

			t.Log("Build job shouldn't be deleted")

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(3))
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionRunning)

			jobList = &batchv1.JobList{}
			err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
			g.Expect(err).To(gomega.BeNil())
			g.Expect(jobList.Items).To(gomega.HaveLen(1))

			t.Log("Service is created")
			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())

			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonServiceCreated)

			svc := &corev1.Service{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, svc)).To(gomega.Succeed())
			g.Expect(err).To(gomega.BeNil())

			g.Expect(svc.Spec.Ports).To(gomega.HaveLen(1))
			g.Expect(svc.Spec.Ports[0].Name).To(gomega.Equal("http"))
			g.Expect(svc.Spec.Ports[0].TargetPort).To(gomega.Equal(intstr.FromInt(8080)))

			g.Expect(isSubset(svc.Spec.Selector, job.Spec.Template.Labels)).
				To(gomega.BeFalse(), "svc selector should not catch job pods")

			g.Expect(svc.Spec.Selector).To(gomega.Equal(deployment.Spec.Selector.MatchLabels))

			t.Log("HPA is created")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beOKReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonHorizontalPodAutoscalerCreated)

			hpaList := &autoscalingv1.HorizontalPodAutoscalerList{}
			err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, hpaList)
			g.Expect(err).To(gomega.BeNil())
			g.Expect(hpaList.Items).To(gomega.HaveLen(1))

			hpaSpec := hpaList.Items[0].Spec

			g.Expect(hpaSpec.ScaleTargetRef.Name).To(gomega.Equal(function.GetName()))
			g.Expect(hpaSpec.ScaleTargetRef.Kind).To(gomega.Equal(serverlessv1alpha2.FunctionKind))
			g.Expect(hpaSpec.ScaleTargetRef.APIVersion).To(gomega.Equal(serverlessv1alpha2.GroupVersion.String()))

			t.Log("Deployment is ready")
			deployment.Status.Conditions = []appsv1.DeploymentCondition{
				{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: MinimumReplicasAvailable},
				{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: NewRSAvailableReason},
			}
			g.Expect(resourceClient.Status().Update(context.TODO(), deployment)).To(gomega.Succeed())

			g.Expect(reconciler.Reconcile(ctx, request)).To(beFinishedReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveConditionRunning)
			g.Expect(function).To(haveConditionReasonDeploymentReady)

			t.Log("should not change state on reconcile")
			g.Expect(reconciler.Reconcile(ctx, request)).To(beFinishedReconcileResult)

			function = &serverlessv1alpha2.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveConditionRunning)
			g.Expect(function).To(haveConditionReasonDeploymentReady)

		})
	}
}

func TestGitOps_GitErrorHandling(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rtm := serverlessv1alpha2.NodeJs18

	resourceClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)

	testCfg := setUpControllerConfig(g)

	initializeServerlessResources(g, resourceClient)

	createDockerfileForRuntime(g, resourceClient, rtm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Check if Requeue is set to true in case of recoverable error", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)

		function := &serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{Name: "git-fn", Namespace: testNamespace},
			Spec: serverlessv1alpha2.FunctionSpec{
				Source: serverlessv1alpha2.Source{
					GitRepository: &serverlessv1alpha2.GitRepositorySource{
						Repository: serverlessv1alpha2.Repository{
							BaseDir:   "dir",
							Reference: "ref",
						},
					},
				},
				Runtime: rtm,
			},
		}

		g.Expect(resourceClient.Create(context.TODO(), function)).To(gomega.Succeed())
		// We don't use MakeGitError2 function because: https://github.com/libgit2/git2go/issues/873
		gitErr := &git2go.GitError{Message: "NotFound", Class: 0, Code: git2go.ErrorCodeNotFound}
		gitOpts := git.Options{URL: "", Reference: "ref"}
		gitClient := &automock.GitClient{}
		gitClient.On("LastCommit", gitOpts).Return("", gitErr)
		defer gitClient.AssertExpectations(t)

		factory := automock.NewGitClientFactory(t)
		factory.On("GetGitClient", mock.Anything).Return(gitClient)
		defer factory.AssertExpectations(t)

		prometheusCollector := &automock.StatsCollector{}
		prometheusCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()
		request := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: function.GetNamespace(),
				Name:      function.GetName(),
			},
		}

		reconciler := &FunctionReconciler{
			Log:               zap.NewNop().Sugar(),
			client:            resourceClient,
			recorder:          record.NewFakeRecorder(100),
			config:            testCfg,
			gitFactory:        factory,
			statsCollector:    prometheusCollector,
			initStateFunction: stateFnGitCheckSources,
		}

		//WHEN
		res, err := reconciler.Reconcile(ctx, request)

		//THEN
		g.Expect(err).To(gomega.BeNil())
		g.Expect(res.Requeue).To(gomega.BeFalse())

		var updatedFn serverlessv1alpha2.Function
		err = resourceClient.Get(context.TODO(), request.NamespacedName, &updatedFn)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(updatedFn.Status.Conditions).To(gomega.HaveLen(1))
		g.Expect(updatedFn.Status.Conditions[0].Message).To(gomega.Equal("Stop reconciliation, reason: NotFound"))
	})
}

func Test_stateFnGitCheckSources(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rtm := serverlessv1alpha2.NodeJs18

	resourceClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)

	testCfg := setUpControllerConfig(g)

	initializeServerlessResources(g, resourceClient)

	createDockerfileForRuntime(g, resourceClient, rtm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//TODO
	t.Run("Check if requeue in-case of non-existing git-repo-cr", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)

		function := &serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{Name: "git-fn", Namespace: testNamespace},
			Spec: serverlessv1alpha2.FunctionSpec{
				Runtime: rtm,
				Source: serverlessv1alpha2.Source{
					GitRepository: &serverlessv1alpha2.GitRepositorySource{
						Repository: serverlessv1alpha2.Repository{
							BaseDir:   "dir",
							Reference: "ref",
						},
					},
				},
			},
		}
		g.Expect(resourceClient.Create(context.TODO(), function)).To(gomega.Succeed())

		prometheusCollector := &automock.StatsCollector{}
		prometheusCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()
		request := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: function.GetNamespace(),
				Name:      function.GetName(),
			},
		}

		gitClient := new(automock.GitClient)
		gitClient.On("LastCommit", mock.Anything).Return("", fmt.Errorf("test error")).Once()
		defer gitClient.AssertExpectations(t)

		factory := automock.NewGitClientFactory(t)
		factory.On("GetGitClient", mock.Anything).Return(gitClient)
		defer factory.AssertExpectations(t)

		reconciler := &FunctionReconciler{
			Log:               zap.NewNop().Sugar(),
			client:            resourceClient,
			recorder:          record.NewFakeRecorder(100),
			gitFactory:        factory,
			config:            testCfg,
			statsCollector:    prometheusCollector,
			initStateFunction: stateFnGitCheckSources,
		}

		//WHEN
		res, err := reconciler.Reconcile(ctx, request)

		//THEN
		g.Expect(err).To(gomega.BeNil())
		// this is expected to be false, because returning an error is enough to requeue
		g.Expect(res.Requeue).To(gomega.BeTrue())
	})
}

func isSubset(subSet, superSet map[string]string) bool {
	if len(superSet) == 0 {
		return true
	}
	for k, v := range subSet {
		value, ok := superSet[k]
		if !ok || value != v {
			return false
		}
	}
	return true
}
