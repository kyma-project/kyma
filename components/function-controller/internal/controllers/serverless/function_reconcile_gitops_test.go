package serverless

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	git2go "github.com/libgit2/git2go/v31"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/automock"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
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

func TestGitOps(t *testing.T) {
	//GIVEN
	var newMockedGitOperator = func(secretName string, credentials map[string]string, auth *string) *automock.GitOperator {
		options := git.Options{
			URL:       "https://mock.repo/kyma/test",
			Reference: "main",
		}

		if auth != nil {
			options.Auth = &git.AuthOptions{
				Type:        git.RepositoryAuthType(*auth),
				Credentials: credentials,
				SecretName:  secretName,
			}
		}

		mock := new(automock.GitOperator)
		mock.On("LastCommit", options).Return("pierwszy-hash", nil)
		options.Reference = "newone"
		mock.On("LastCommit", options).Return("a376218bdcd705cc39aa7ce7f310769fab6d51c9", nil)

		return mock
	}

	g := gomega.NewGomegaWithT(t)
	rtm := serverlessv1alpha1.Nodejs12
	resourceClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	testCfg := setUpControllerConfig(g)
	initializeServerlessResources(g, resourceClient)
	createDockerfileForRuntime(g, resourceClient, rtm)

	for _, testData := range testDataScenarios {
		t.Run(fmt.Sprintf("[%s] should successfully update Function]", testData.info), func(t *testing.T) {
			//GIVEN
			g := gomega.NewGomegaWithT(t)
			inFunction := newTestGitFunction(testNamespace, "ah-tak-przeciez", 1, 2)
			g.Expect(resourceClient.Create(context.TODO(), inFunction)).To(gomega.Succeed())

			var auth *serverlessv1alpha1.RepositoryAuth
			if testData.authType != nil {
				auth = &serverlessv1alpha1.RepositoryAuth{
					Type:       serverlessv1alpha1.RepositoryAuthType(*testData.authType),
					SecretName: inFunction.Name,
				}

				secret := newTestSecret(inFunction.Name, testNamespace, testData.stringData)
				g.Expect(resourceClient.Create(context.TODO(), secret)).To(gomega.Succeed())
			}

			repo := newTestRepository(inFunction.GetName(), testNamespace, auth)
			g.Expect(resourceClient.Create(context.TODO(), repo)).To(gomega.Succeed())

			request := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: inFunction.GetNamespace(),
					Name:      inFunction.GetName(),
				},
			}
			operator := newMockedGitOperator(inFunction.Name, testData.stringData, testData.authType)
			statsCollector := &automock.StatsCollector{}
			statsCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()

			reconciler := &FunctionReconciler{
				Log:            log.Log,
				client:         resourceClient,
				recorder:       record.NewFakeRecorder(100),
				config:         testCfg,
				gitOperator:    operator,
				statsCollector: statsCollector,
			}

			fnLabels := reconciler.internalFunctionLabels(inFunction)

			//WHEN
			t.Log("creating the Function")
			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)
			// verify function
			function := &serverlessv1alpha1.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(1))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonSourceUpdated)

			t.Log("creating the Job")
			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			function = &serverlessv1alpha1.Function{}
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
			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			function = &serverlessv1alpha1.Function{}
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

			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			function = &serverlessv1alpha1.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonJobFinished)

			t.Log("change function branch")
			function.Spec.Reference = "newone"
			g.Expect(resourceClient.Update(context.TODO(), function)).To(gomega.Succeed())

			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			// check if status was updated
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveStatusReference("newone"))
			g.Expect(function).To(haveStatusCommit("a376218bdcd705cc39aa7ce7f310769fab6d51c9"))

			t.Log("delete the old Job")
			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			function = &serverlessv1alpha1.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveUnknownConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)

			jobList = &batchv1.JobList{}
			err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
			g.Expect(err).To(gomega.BeNil())
			g.Expect(jobList.Items).To(gomega.HaveLen(0))

			t.Log("creating the Job")
			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			function = &serverlessv1alpha1.Function{}
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
			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			function = &serverlessv1alpha1.Function{}
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

			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			function = &serverlessv1alpha1.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(2))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveUnknownConditionRunning)
			g.Expect(function).To(haveConditionReasonJobFinished)

			t.Log("deploy started")
			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			function = &serverlessv1alpha1.Function{}
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
			expectedImage := reconciler.buildImageAddress(function, "registry.kyma.local")
			g.Expect(deployment).To(gomega.Not(gomega.BeNil()))
			g.Expect(deployment).To(haveSpecificContainer0Image(expectedImage))
			g.Expect(deployment).To(haveLabelLen(7))
			g.Expect(deployment).To(haveLabelWithValue(serverlessv1alpha1.FunctionNameLabel, function.Name))
			g.Expect(deployment).To(haveLabelWithValue(serverlessv1alpha1.FunctionManagedByLabel, serverlessv1alpha1.FunctionControllerValue))
			g.Expect(deployment).To(haveLabelWithValue(serverlessv1alpha1.FunctionUUIDLabel, string(function.UID)))
			g.Expect(deployment).To(haveLabelWithValue(
				serverlessv1alpha1.FunctionResourceLabel, serverlessv1alpha1.FunctionResourceLabelDeploymentValue))

			g.Expect(deployment).To(haveLabelWithValue(testBindingLabel1, "foobar"))
			g.Expect(deployment).To(haveLabelWithValue(testBindingLabel2, testBindingLabelValue))
			g.Expect(deployment).To(haveLabelWithValue("foo", "bar"))

			t.Log("service creation")
			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			function = &serverlessv1alpha1.Function{}
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

			g.Expect(labels.AreLabelsInWhiteList(svc.Spec.Selector, job.Spec.Template.Labels)).
				To(gomega.BeFalse(), "svc selector should not catch job pods")

			g.Expect(svc.Spec.Selector).To(gomega.Equal(deployment.Spec.Selector.MatchLabels))

			t.Log("hpa creation")
			g.Expect(reconciler.Reconcile(request)).To(beOKReconcileResult)

			function = &serverlessv1alpha1.Function{}
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

			g.Expect(hpaSpec.ScaleTargetRef.Name).To(gomega.Equal(deployment.GetName()))
			g.Expect(hpaSpec.ScaleTargetRef.Kind).To(gomega.Equal("Deployment"))
			g.Expect(hpaSpec.ScaleTargetRef.APIVersion).To(gomega.Equal(appsv1.SchemeGroupVersion.String()))

			t.Log("deployment ready")
			deployment.Status.Conditions = []appsv1.DeploymentCondition{
				{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: MinimumReplicasAvailable},
				{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: NewRSAvailableReason},
			}
			g.Expect(resourceClient.Status().Update(context.TODO(), deployment)).To(gomega.Succeed())

			g.Expect(reconciler.Reconcile(request)).To(beFinishedReconcileResult)

			function = &serverlessv1alpha1.Function{}
			g.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
			g.Expect(function).To(haveConditionLen(conditionLen))
			g.Expect(function).To(haveConditionCfgRdy)
			g.Expect(function).To(haveConditionBuildRdy)
			g.Expect(function).To(haveConditionRunning)
			g.Expect(function).To(haveConditionReasonDeploymentReady)

			t.Log("should not change state on reconcile")
			g.Expect(reconciler.Reconcile(request)).To(beFinishedReconcileResult)

			function = &serverlessv1alpha1.Function{}
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
	testRepoName := "test-repo"
	rtm := serverlessv1alpha1.Nodejs12
	resourceClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	testCfg := setUpControllerConfig(g)
	initializeServerlessResources(g, resourceClient)
	createDockerfileForRuntime(g, resourceClient, rtm)
	t.Run("Check if Requeue is set to true in case of recoverable error", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		gitRepo := serverlessv1alpha1.GitRepository{
			ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: testNamespace},
			Spec:       serverlessv1alpha1.GitRepositorySpec{},
		}
		err := resourceClient.Create(context.TODO(), &gitRepo)
		g.Expect(err).To(gomega.BeNil())

		function := &serverlessv1alpha1.Function{
			ObjectMeta: metav1.ObjectMeta{Name: "git-fn", Namespace: testNamespace},
			Spec: serverlessv1alpha1.FunctionSpec{
				Source:     testRepoName,
				Runtime:    rtm,
				Type:       serverlessv1alpha1.SourceTypeGit,
				Repository: serverlessv1alpha1.Repository{BaseDir: "dir", Reference: "ref"},
			},
		}
		g.Expect(resourceClient.Create(context.TODO(), function)).To(gomega.Succeed())
		// We don't use MakeGitError2 function because: https://github.com/libgit2/git2go/issues/873
		gitErr := &git2go.GitError{Message: "NotFound", Class: 0, Code: git2go.ErrorCodeNotFound}
		gitOpts := git.Options{URL: "", Reference: "ref"}
		operator := &automock.GitOperator{}
		operator.On("LastCommit", gitOpts).Return("", gitErr)
		defer operator.AssertExpectations(t)

		prometheusCollector := &automock.StatsCollector{}
		prometheusCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()
		request := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: function.GetNamespace(),
				Name:      function.GetName(),
			},
		}

		reconciler := &FunctionReconciler{
			Log:            log.Log,
			client:         resourceClient,
			recorder:       record.NewFakeRecorder(100),
			config:         testCfg,
			gitOperator:    operator,
			statsCollector: prometheusCollector,
		}

		//WHEN
		res, err := reconciler.Reconcile(request)

		//THEN
		g.Expect(err).To(gomega.BeNil())
		g.Expect(res.Requeue).To(gomega.BeFalse())

		var updatedFn serverlessv1alpha1.Function
		err = resourceClient.Get(context.TODO(), request.NamespacedName, &updatedFn)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(updatedFn.Status.Conditions).To(gomega.HaveLen(1))
		g.Expect(updatedFn.Status.Conditions[0].Message).To(gomega.Equal("Stop reconciliation, reason: NotFound"))
	})
}
