package controllers

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

var _ = ginkgo.Describe("Subscription", func() {
	var (
		request    ctrl.Request
		cfg        *rest.Config
		k8sClient  client.Client // You'll be using this client in your tests.
		testEnv    *envtest.Environment
		reconciler *SubscriptionReconciler
	)

	ginkgo.BeforeEach(func() {
		// Prepare test fixtures
		subscription := eventingv1alpha1.Subscription{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       eventingv1alpha1.SubscriptionSpec{},
			Status:     eventingv1alpha1.SubscriptionStatus{},
		}
		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: subscription.GetNamespace(), Name: subscription.GetName()}}
		reconciler := SubscriptionReconciler{
			Client: resource.Client,
			Log:    nil,
			Scheme: nil,
		}

		// Start k8s test cluster
		cfg, err := testEnv.Start()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		gomega.Expect(cfg).ToNot(gomega.BeNil())

		// Add Subscription scheme
		err = eventingv1alpha1.AddToScheme(scheme.Scheme)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		// "This marker is what allows new schemas to be added here automatically when a new API is added to the project."
		// Source: https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html
		// +kubebuilder:scaffold:scheme

		// Create k8s client
		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		gomega.Expect(k8sClient).ToNot(gomega.BeNil())
	})

	ginkgo.It("should create Subscription", func() {
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
	})

})
