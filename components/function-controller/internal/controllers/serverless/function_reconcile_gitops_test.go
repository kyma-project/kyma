package serverless

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/automock"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
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
)

type data struct {
	authType   *string
	stringData map[string]string
}

var (
	authTypeBasic = "basic"
	authTypeKey   = "key"
)

var testData = []data{
	{
		authType: &authTypeKey,
		stringData: map[string]string{
			"user":     "test",
			"password": "test",
		},
	},
	{
		authType: &authTypeBasic,
		stringData: map[string]string{
			"user":     "test",
			"password": "test",
		},
	},
	{
		authType: &authTypeKey,
		stringData: map[string]string{
			"authTypeKey": "123",
		},
	},
	{
		authType:   nil,
		stringData: nil,
	},
}

var _ = ginkgo.Describe("Function", func() {
	var newTestFunction = func(namespace, name string, minReplicas, maxReplicas int) *serverlessv1alpha1.Function {
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
				Runtime:    serverlessv1alpha1.Nodejs12,
				Repository: serverlessv1alpha1.Repository{
					BaseDir:   "/",
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

	var newTestRepository = func(name, namespace string, auth *serverlessv1alpha1.RepositoryAuth) *serverlessv1alpha1.GitRepository {
		return &serverlessv1alpha1.GitRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: serverlessv1alpha1.GitRepositorySpec{
				URL:  "https://mock.repo/kyma/test",
				Auth: auth,
			},
		}
	}

	// creates secret required for authentication
	var newTestSecret = func(name, namespace string, stringData map[string]string) *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			StringData: stringData,
		}
	}

	var newMockedGitOperator = func(secretName string, credentials map[string]string, auth *string) *automock.GitOperator {
		options := git.Options{
			URL:       "https://mock.repo/kyma/test",
			Reference: "master",
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

	for _, data := range testData {
		var (
			reconciler *FunctionReconciler
			request    ctrl.Request
			fnLabels   map[string]string
		)

		ginkgo.Context("gitops function", func() {

			data := data
			ginkgo.BeforeEach(func() {
				function := newTestFunction(testNamespace, "ah-tak-przeciez", 1, 2)

				request = ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: function.GetNamespace(),
						Name:      function.GetName(),
					},
				}

				var auth *serverlessv1alpha1.RepositoryAuth
				var operator *automock.GitOperator

				if data.authType != nil {
					auth = &serverlessv1alpha1.RepositoryAuth{
						Type:       serverlessv1alpha1.RepositoryAuthType(*data.authType),
						SecretName: function.Name,
					}

					secret := newTestSecret(function.Name, testNamespace, data.stringData)
					// apply secret for a given scenario
					gomega.Expect(resourceClient.Create(context.TODO(), secret)).To(gomega.Succeed())
				}

				operator = newMockedGitOperator(function.Name, data.stringData, data.authType)

				repo := newTestRepository(function.GetName(), testNamespace, auth)

				// apply git repository for a given scenario
				gomega.Expect(resourceClient.Create(context.TODO(), repo)).To(gomega.Succeed())

				// apply function for a given scenario
				gomega.Expect(resourceClient.Create(context.TODO(), function)).To(gomega.Succeed())

				reconciler = &FunctionReconciler{
					Log:         log.Log,
					client:      resourceClient,
					recorder:    record.NewFakeRecorder(100),
					config:      config,
					gitOperator: operator,
				}

				fnLabels = reconciler.internalFunctionLabels(function)
			})

			ginkgo.It("should successfully update Function", func() {
				ginkgo.By("creating the Function")
				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)
				// verify function
				function := &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(1),
					haveConditionCfgRdy,
					haveUnknownConditionBuildRdy,
					haveUnknownConditionRunning,
					haveConditionReasonSourceUpdated))

				ginkgo.By("creating the Job")
				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(2),
					haveConditionCfgRdy,
					haveUnknownConditionBuildRdy,
					haveUnknownConditionRunning))

				jobList := &batchv1.JobList{}
				err := reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(jobList.Items).To(gomega.HaveLen(1))

				ginkgo.By("build in progress")
				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(2),
					haveConditionCfgRdy,
					haveUnknownConditionBuildRdy,
					haveUnknownConditionRunning,
					haveConditionReasonJobRunning,
				))

				ginkgo.By("build finished")
				job := &batchv1.Job{}
				gomega.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{
					Namespace: jobList.Items[0].GetNamespace(),
					Name:      jobList.Items[0].GetName(),
				}, job)).To(gomega.Succeed())

				gomega.Expect(job).ToNot(gomega.BeNil())
				job.Status.Succeeded = 1
				now := metav1.Now()
				job.Status.CompletionTime = &now
				gomega.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(2),
					haveConditionCfgRdy,
					haveConditionBuildRdy,
					haveUnknownConditionRunning,
					haveConditionReasonJobFinished))

				ginkgo.By("change function branch")
				function.Spec.Reference = "newone"
				gomega.Expect(resourceClient.Update(context.TODO(), function)).To(gomega.Succeed())

				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				// check if status was updated
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(2),
					haveConditionCfgRdy,
					haveConditionBuildRdy,
					haveUnknownConditionRunning,
					haveStatusReference("newone"),
					haveStatusCommit("a376218bdcd705cc39aa7ce7f310769fab6d51c9"),
				))

				ginkgo.By("delete the old Job")
				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(2),
					haveConditionCfgRdy,
					haveUnknownConditionBuildRdy,
					haveUnknownConditionRunning))

				jobList = &batchv1.JobList{}
				err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(jobList.Items).To(gomega.HaveLen(0))

				ginkgo.By("creating the Job")
				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(2),
					haveConditionCfgRdy,
					haveUnknownConditionBuildRdy,
					haveUnknownConditionRunning))

				jobList = &batchv1.JobList{}
				err = reconciler.client.ListByLabel(context.TODO(), function.GetNamespace(), fnLabels, jobList)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(jobList.Items).To(gomega.HaveLen(1))

				ginkgo.By("build in progress")
				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(2),
					haveConditionCfgRdy,
					haveUnknownConditionBuildRdy,
					haveUnknownConditionRunning,
					haveConditionReasonJobRunning))

				ginkgo.By("build finished")
				job = &batchv1.Job{}
				gomega.Expect(resourceClient.Get(context.TODO(), types.NamespacedName{
					Namespace: jobList.Items[0].GetNamespace(),
					Name:      jobList.Items[0].GetName(),
				}, job)).To(gomega.Succeed())
				gomega.Expect(job).ToNot(gomega.BeNil())
				job.Status.Succeeded = 1
				now = metav1.Now()
				job.Status.CompletionTime = &now
				gomega.Expect(resourceClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(2),
					haveConditionCfgRdy,
					haveConditionBuildRdy,
					haveUnknownConditionRunning,
					haveConditionReasonJobFinished))

				ginkgo.By("deploy started")
				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(conditionLen),
					haveConditionCfgRdy,
					haveConditionBuildRdy,
					haveUnknownConditionRunning))

				deployments := &appsv1.DeploymentList{}
				gomega.Expect(resourceClient.ListByLabel(context.TODO(), request.Namespace, fnLabels, deployments)).To(gomega.Succeed())
				gomega.Expect(len(deployments.Items)).To(gomega.Equal(1))

				deployment := &deployments.Items[0]
				expectedImage := reconciler.buildImageAddress(function)
				gomega.Expect(deployment).To(gomega.And(
					gomega.Not(gomega.BeNil()),
					haveSpecificContainer0Image(expectedImage),
					haveLabelLen(7),
					haveLabelWithValue(serverlessv1alpha1.FunctionNameLabel, function.Name),
					haveLabelWithValue(serverlessv1alpha1.FunctionManagedByLabel, "function-controller"),
					haveLabelWithValue(serverlessv1alpha1.FunctionUUIDLabel, string(function.UID)),
					haveLabelWithValue(serverlessv1alpha1.FunctionResourceLabel, serverlessv1alpha1.FunctionResourceLabelDeploymentValue),
					haveLabelWithValue(testBindingLabel1, "foobar"),
					haveLabelWithValue(testBindingLabel2, testBindingLabelValue),
					haveLabelWithValue("foo", "bar"),
				))

				ginkgo.By("service creation")
				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(conditionLen),
					haveConditionCfgRdy,
					haveConditionBuildRdy,
					haveUnknownConditionRunning,
					haveConditionReasonServiceCreated))

				svc := &corev1.Service{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, svc)).To(gomega.Succeed())
				gomega.Expect(err).To(gomega.BeNil())

				gomega.Expect(svc.Spec.Ports).To(gomega.HaveLen(1))
				gomega.Expect(svc.Spec.Ports[0].Name).To(gomega.Equal("http"))
				gomega.Expect(svc.Spec.Ports[0].TargetPort).To(gomega.Equal(intstr.FromInt(8080)))

				gomega.Expect(labels.AreLabelsInWhiteList(svc.Spec.Selector, job.Spec.Template.Labels)).
					To(gomega.BeFalse(), "svc selector should not catch job pods")

				gomega.Expect(svc.Spec.Selector).To(gomega.Equal(deployment.Spec.Selector.MatchLabels))

				ginkgo.By("hpa creation")
				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(conditionLen),
					haveConditionCfgRdy,
					haveConditionBuildRdy,
					haveUnknownConditionRunning,
					haveConditionReasonHorizontalPodAutoscalerCreated))

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

				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(conditionLen),
					haveConditionCfgRdy,
					haveConditionBuildRdy,
					haveConditionRunning,
					haveConditionReasonDeploymentReady))

				ginkgo.By("should not change state on reconcile")
				gomega.Ω(reconciler.Reconcile(request)).To(beOKReconcileResult)

				function = &serverlessv1alpha1.Function{}
				gomega.Expect(resourceClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
				gomega.Expect(function).To(gomega.And(
					haveConditionLen(conditionLen),
					haveConditionCfgRdy,
					haveConditionBuildRdy,
					haveConditionRunning,
					haveConditionReasonDeploymentReady))
			})
		})
	}
})
