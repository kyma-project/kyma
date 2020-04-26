package kubernetes

import (
	"context"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = ginkgo.Describe("ServiceAccount", func() {
	var (
		reconciler         *ServiceAccountReconciler
		request            ctrl.Request
		baseServiceAccount *corev1.ServiceAccount
		namespace          string
	)

	ginkgo.BeforeEach(func() {
		userNamespace := newFixNamespace("tam")
		gomega.Expect(resourceClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

		baseServiceAccount = newFixBaseServiceAccount(config.BaseNamespace, "ah-tak-przeciez")
		gomega.Expect(resourceClient.Create(context.TODO(), baseServiceAccount)).To(gomega.Succeed())

		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseServiceAccount.GetNamespace(), Name: baseServiceAccount.GetName()}}
		reconciler = NewServiceAccount(k8sClient, log.Log, config, serviceAccountSvc)
		namespace = userNamespace.GetName()
	})

	ginkgo.It("should successfully propagate base ServiceAccount to user namespace", func() {
		ginkgo.By("reconciling the Secret")
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.ServiceAccountRequeueDuration))

		serviceAccount := &corev1.ServiceAccount{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(serviceAccount, baseServiceAccount)

		ginkgo.By("updating the base ServiceAccount")
		copy := baseServiceAccount.DeepCopy()
		copy.Labels["test"] = "value"
		copy.AutomountServiceAccountToken = nil
		gomega.Expect(k8sClient.Update(context.TODO(), copy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.ServiceAccountRequeueDuration))

		serviceAccount = &corev1.ServiceAccount{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(serviceAccount, copy)

		ginkgo.By("updating the modified ServiceAccount in user namespace")
		userCopy := serviceAccount.DeepCopy()
		trueValue := true
		userCopy.AutomountServiceAccountToken = &trueValue
		gomega.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.ServiceAccountRequeueDuration))

		serviceAccount = &corev1.ServiceAccount{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(serviceAccount, copy)
	})
})
