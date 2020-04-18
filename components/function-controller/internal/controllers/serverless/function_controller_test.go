package serverless

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var _ = ginkgo.Describe("Function", func() {
	var (
		reconciler *FunctionReconciler
		request    ctrl.Request
	)

	ginkgo.BeforeEach(func() {
		function := newFixFunction("tutaj", "ah-tak-przeciez")
		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: function.GetNamespace(), Name: function.GetName()}}
		gomega.Expect(k8sClient.Create(context.TODO(), function)).To(gomega.Succeed())

		reconciler = NewFunction(k8sClient, log.Log, config, scheme.Scheme, record.NewFakeRecorder(100))
	})

	ginkgo.It("should successfully create Function", func() {
		ginkgo.By("creating the ConfigMap")
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function := &serverlessv1alpha1.Function{}
		gomega.Expect(k8sClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(1))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		configMapList := &corev1.ConfigMapList{}
		err = reconciler.resourceClient.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.functionLabels(function), configMapList)
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
		gomega.Expect(k8sClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(2))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		jobList := &batchv1.JobList{}
		err = reconciler.resourceClient.ListByLabel(context.TODO(), function.GetNamespace(), reconciler.functionLabels(function), jobList)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(jobList.Items).To(gomega.HaveLen(1))

		ginkgo.By("build in progress")
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(k8sClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(2))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionUnknown))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		ginkgo.By("build finished")
		job := &batchv1.Job{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: jobList.Items[0].GetNamespace(), Name: jobList.Items[0].GetName()}, job)).To(gomega.Succeed())
		gomega.Expect(job).ToNot(gomega.BeNil())
		job.Status.Succeeded = 1
		now := metav1.Now()
		job.Status.CompletionTime = &now
		gomega.Expect(k8sClient.Status().Update(context.TODO(), job)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(k8sClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(2))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		ginkgo.By("deploy started")
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(k8sClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(3))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionUnknown))

		service := &servingv1.Service{}
		gomega.Expect(k8sClient.Get(context.TODO(), request.NamespacedName, service)).To(gomega.Succeed())
		gomega.Expect(service).ToNot(gomega.BeNil())
		gomega.Expect(service.Spec.Template.Spec.Containers).To(gomega.HaveLen(1))
		gomega.Expect(service.Spec.Template.Spec.Containers[0].Image).To(gomega.Equal(reconciler.buildExternalImageAddress(function)))

		ginkgo.By("running")
		service.Status.Conditions = duckv1.Conditions{{Type: apis.ConditionReady, Status: corev1.ConditionTrue}}
		gomega.Expect(k8sClient.Status().Update(context.TODO(), service)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.RequeueDuration))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(k8sClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(3))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))

		ginkgo.By("after status update")
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.RequeueDuration))

		function = &serverlessv1alpha1.Function{}
		gomega.Expect(k8sClient.Get(context.TODO(), request.NamespacedName, function)).To(gomega.Succeed())
		gomega.Expect(function.Status.Conditions).To(gomega.HaveLen(3))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionConfigurationReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionBuildReady)).To(gomega.Equal(corev1.ConditionTrue))
		gomega.Expect(reconciler.getConditionStatus(function.Status.Conditions, serverlessv1alpha1.ConditionRunning)).To(gomega.Equal(corev1.ConditionTrue))
	})

	ginkgo.It("should handle reconcilation lags", func() {
		ginkgo.By("handling not existing Function")
		result, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "nope", Name: "noooooopppeee"}})
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))
	})
})

func newFixFunction(namespace, name string) *serverlessv1alpha1.Function {
	one := int32(1)
	two := int32(2)
	suffix := rand.Int()
	return &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", name, suffix),
			Namespace: namespace,
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			Source: "module.exports = {main: function(event, context) {return 'Hello World 321'}}",
			Deps:   "   ",
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
		},
	}
}

