package config

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeout  = time.Second * 20
	interval = time.Second * 1
)

var _ = Describe("Config Controller", func() {
	Context("should watch for namespace creation and", func() {
		//It("create Runtimes, Credentials and ServiceAccount if the namespace is not a system one", func() {
		//	testNamespaceName := "epsteindidntkillhimself"
		//	expectedRequest := &reconcile.Request{NamespacedName: types.NamespacedName{Name: testNamespaceName}}
		//
		//	t := setupTest(NamespaceType)
		//	testNamespace := &corev1.Namespace{
		//		ObjectMeta: v1.ObjectMeta{
		//			Name: testNamespaceName,
		//		},
		//	}
		//
		//	err := t.c.Create(context.TODO(), testNamespace)
		//	Expect(err).NotTo(HaveOccurred())
		//	Eventually(t.requests, timeout).Should(Receive(Equal(*expectedRequest)))
		//
		//	t.c.Delete(context.TODO(), testNamespace)
		//	close(t.stopMgr)
		//	t.mgrStopped.Wait()
		//})

		It("not create a Runtimes, Credentials and ServiceAccount if the namespace is a system one", func() {
			testNamespaceName := excludedNamespace1
			testNamespace := &corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: testNamespaceName,
				},
			}

			err := k8sClient.Create(context.Background(), testNamespace)
			Expect(err).NotTo(HaveOccurred())

			//runtime := corev1.ConfigMap{}
			//key := client.ObjectKey{Name: runtimeName, Namespace: testNamespaceName}
			//err = k8sClient.Get(context.Background(), key, &runtime)
			//Expect(err).NotTo(HaveOccurred())

			Eventually(func() error {
				namespace := corev1.Namespace{}
				key := client.ObjectKey{Name: testNamespaceName, Namespace: testNamespaceName}
				return k8sClient.Get(context.Background(), key, &namespace)
			}, timeout, interval).ShouldNot(HaveOccurred())

			// Verify if the Runtimes, Credentials and ServiceAccount are not created
			Eventually(func() error {
				runtime := corev1.ConfigMap{}
				key := client.ObjectKey{Name: runtimeName, Namespace: testNamespaceName}
				return k8sClient.Get(context.Background(), key, &runtime)
			}, timeout, interval).Should(HaveOccurred())

			Eventually(func() error {
				credential := corev1.Secret{}
				key := client.ObjectKey{Name: credentialName, Namespace: testNamespaceName}
				return k8sClient.Get(context.Background(), key, &credential)
			}, timeout, interval).Should(HaveOccurred())

			Eventually(func() error {
				serviceAccount := corev1.ServiceAccount{}
				key := client.ObjectKey{Name: serviceAccountName, Namespace: testNamespaceName}
				return k8sClient.Get(context.Background(), key, &serviceAccount)
			}, timeout, interval).Should(HaveOccurred())
		})
	})
})
