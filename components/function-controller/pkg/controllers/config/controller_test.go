package config

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config Controller", func() {
	const timeout = time.Millisecond * 250

	Context("should watch for namespace creation and", func() {
		It("create Runtimes, Credentials and ServiceAccount if the namespace is not a system one", func() {
			_, err := getConfigMap(runtimeName, includedNamespace)
			Expect(err).To(HaveOccurred())

			_, err = getSecret(registryCredentialName, includedNamespace)
			Expect(err).To(HaveOccurred())

			_, err = getSecret(imagePullSecretName, includedNamespace)
			Expect(err).To(HaveOccurred())

			_, err = getServiceAccount(serviceAccountName, includedNamespace)
			Expect(err).To(HaveOccurred())

			namespace := fixNamespace(includedNamespace)

			// Create namespace
			Expect(k8sClient.Create(context.Background(), namespace)).Should(Succeed())
			time.Sleep(timeout)

			_, err = getConfigMap(runtimeName, includedNamespace)
			Expect(err).NotTo(HaveOccurred())

			_, err = getSecret(registryCredentialName, includedNamespace)
			Expect(err).NotTo(HaveOccurred())

			_, err = getSecret(imagePullSecretName, includedNamespace)
			Expect(err).NotTo(HaveOccurred())

			_, err = getServiceAccount(serviceAccountName, includedNamespace)
			Expect(err).NotTo(HaveOccurred())
		})

		It("not create a Runtimes, Credentials and ServiceAccount if the namespace is a system one", func() {
			namespace := fixNamespace(excludedNamespace)

			// Create namespace
			Expect(k8sClient.Create(context.Background(), namespace)).Should(Succeed())
			time.Sleep(timeout)

			_, err := getConfigMap(runtimeName, excludedNamespace)
			Expect(err).To(HaveOccurred())

			_, err = getSecret(registryCredentialName, excludedNamespace)
			Expect(err).To(HaveOccurred())

			_, err = getServiceAccount(serviceAccountName, excludedNamespace)
			Expect(err).To(HaveOccurred())
		})

		It("propagate changes in base Credentials and Runtimes to not system namespaces", func() {
			// Runtime
			runtimeBeforeUpdate := fixRuntime(runtimeName, baseNamespace, runtimeLabel, nil)

			// Create runtime - we must firstly create runtime in manager client, then we use fake clientset
			Expect(k8sClient.Create(context.Background(), runtimeBeforeUpdate)).Should(Succeed())
			time.Sleep(timeout)

			runtimeAfterUpdate := fixRuntime(runtimeName, baseNamespace, runtimeLabel, map[string]string{
				"foo": "bar",
			})

			// Update runtime
			Expect(k8sClient.Update(context.Background(), runtimeAfterUpdate)).Should(Succeed())
			time.Sleep(4 * timeout)

			r, err := getConfigMap(runtimeName, includedNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(r.Labels["foo"]).To(Equal("bar"))

			// Credential
			credentialBeforeUpdate := fixCredential(registryCredentialName, baseNamespace, registryCredentialName, nil)

			// Create credential - we must firstly create credential in manager client, then we use fake clientset
			Expect(k8sClient.Create(context.Background(), credentialBeforeUpdate)).Should(Succeed())
			time.Sleep(timeout)

			credentialAfterUpdate := fixCredential(registryCredentialName, baseNamespace, registryCredentialName, map[string]string{
				"foo": "bar",
			})

			// Update runtime
			Expect(k8sClient.Update(context.Background(), credentialAfterUpdate)).Should(Succeed())
			time.Sleep(4 * timeout)

			c, err := getSecret(registryCredentialName, includedNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(c.Labels["foo"]).To(Equal("bar"))
		})

		It("recover base Credentials and Runtimes in not system namespaces after deletion", func() {
			err := deleteConfigMap(runtimeName, includedNamespace)
			Expect(err).NotTo(HaveOccurred())

			err = deleteSecret(registryCredentialName, includedNamespace)
			Expect(err).NotTo(HaveOccurred())

			// Wait 3 second to recovery
			time.Sleep(12 * timeout)

			r, err := getConfigMap(runtimeName, includedNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(r.Labels["foo"]).To(Equal("bar"))

			c, err := getSecret(registryCredentialName, includedNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(c.Labels["foo"]).To(Equal("bar"))
		})
	})
})
