/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package function

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/log"

	knapis "knative.dev/pkg/apis"

	fnapis "github.com/kyma-project/kyma/components/function-controller/pkg/apis"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"

	"github.com/kyma-project/kyma/tests/function-controller/framework/function"
	"github.com/kyma-project/kyma/tests/function-controller/framework/registry"
	"github.com/kyma-project/kyma/tests/function-controller/framework/taskrun"
)

var (
	functionGVR = serverlessv1alpha1.SchemeGroupVersion.WithResource("functions")
	taskrunGVR  = tektonv1alpha1.SchemeGroupVersion.WithResource("taskruns")
)

func init() {
	sb := runtime.NewSchemeBuilder(
		fnapis.AddToScheme,
		tektonv1alpha1.AddToScheme,
	)
	if err := sb.AddToScheme(scheme.Scheme); err != nil {
		framework.Failf("Error adding custom resources to Scheme: %v", err)
	}
}

var _ = Describe("Functions", func() {
	var (
		ns  string
		err error

		fnCli      dynamic.ResourceInterface
		taskrunCli dynamic.ResourceInterface

		registryURL string
	)

	f := framework.NewDefaultFramework("function")

	BeforeEach(func() {
		ns = f.Namespace.Name
		fnCli = f.DynamicClient.Resource(functionGVR).Namespace(ns)
		taskrunCli = f.DynamicClient.Resource(taskrunGVR).Namespace(ns)

		By("creating Dockerfiles")

		function.CreateDockerfiles(f)

		By("deploying local container registry")

		registryURL = registry.DeployLocal(f)
		log.Logf("Local container registry URL: %s", registryURL)

		By("configuring controller to use local registry")

		framework.AddCleanupAction(
			// TODO(antoineco): move to global 'Before' func to avoid
			// conflicts if we enable parallelism
			function.SetControllerRegistry(f.ClientSet, registryURL),
		)

		By("creating build ServiceAccount")

		saName := function.ServiceAccountName(f.ClientSet)
		_, err := f.CreateItems(newServiceAccount(saName))
		Expect(err).NotTo(HaveOccurred(), "creating ServiceAccount")
	})

	It("should be able to build and serve", func() {
		// FIXME(antoineco): pulling images fails with "Connection
		// reset by peer" on k8s <1.15 (kubernetes/kubernetes#74840)
		framework.SkipUnlessServerVersionGTE(version.MustParseGeneric("1.15.0"),
			f.ClientSet.Discovery())

		By("creating a valid Function object")

		fn := function.New(ns, "test-helloworld-",
			function.WithDefaults(),
		)

		fnUnstr := function.ToUnstructured(fn)

		fnUnstr, err = fnCli.Create(fnUnstr, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred(), "creating Function object")
		log.Logf("Created Function %q", fnUnstr.GetName())

		By("observing status transitions")

		function.WaitForStatusTransitions(fnUnstr, fnCli, []serverlessv1alpha1.FunctionCondition{
			serverlessv1alpha1.FunctionConditionBuilding,
			// FIXME(antoineco): Function transitions to 'Deploying' before the build completes
			serverlessv1alpha1.FunctionConditionDeploying,
			serverlessv1alpha1.FunctionConditionRunning,
		})

		By("confirming the build TaskRun succeeded")

		trList, err := taskrunCli.List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "listing TaskRuns")
		Expect(trList.Items).To(HaveLen(1), "a unique TaskRun is created for the Function")

		tr := taskrun.FromUnstructured(&trList.Items[0])
		successCond := tr.Status.GetCondition(knapis.ConditionSucceeded)
		Expect(successCond).NotTo(BeNil())
		Expect(successCond.Reason).To(Equal(taskrun.ReasonSucceeded))
	})

	It("should have an error status when the build fails", func() {
		By("creating a Function object with invalid dependencies")

		fn := function.New(ns, "test-invaliddeps-",
			function.WithDefaults(),
		)
		fn.Spec.Deps = "{ invalid }"

		fnUnstr := function.ToUnstructured(fn)

		fnUnstr, err = fnCli.Create(fnUnstr, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred(), "creating Function object")
		log.Logf("Created Function %q", fnUnstr.GetName())

		By("observing status transitions")

		function.WaitForStatusTransitions(fnUnstr, fnCli, []serverlessv1alpha1.FunctionCondition{
			serverlessv1alpha1.FunctionConditionBuilding,
			serverlessv1alpha1.FunctionConditionError,
		})

		By("confirming the build TaskRun failed")

		trList, err := taskrunCli.List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "listing TaskRuns")
		Expect(trList.Items).To(HaveLen(1), "a unique TaskRun is created for the Function")

		tr := taskrun.FromUnstructured(&trList.Items[0])
		successCond := tr.Status.GetCondition(knapis.ConditionSucceeded)
		Expect(successCond).NotTo(BeNil())
		Expect(successCond.Reason).To(Equal(taskrun.ReasonFailed))
	})
})

var _ = Describe("Function webhook", func() {
	var ns string
	var fnCli dynamic.ResourceInterface

	f := framework.NewDefaultFramework("function-webhook")

	BeforeEach(func() {
		ns = f.Namespace.Name
		fnCli = f.DynamicClient.Resource(functionGVR).Namespace(ns)
	})

	It("should reject invalid specs", func() {
		By("creating a Function object with an unknown runtime")

		fn := function.New(ns, "test-invalidspec-")
		fn.Spec.Runtime = "invalid"

		fnUnstr := function.ToUnstructured(fn)

		_, err := fnCli.Create(fnUnstr, metav1.CreateOptions{})
		Expect(err).To(HaveOccurred(), "webhook rejects Function creation")
		Expect(err.Error()).To(ContainSubstring("runtime should be one of"), "enumerates supported runtimes")
	})

	It("should set defaults", func() {
		By("creating a Function object without runtime")

		fn := function.New(ns, "test-missingspec-")

		fnUnstr := function.ToUnstructured(fn)

		fnUnstr, err := fnCli.Create(fnUnstr, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred(), "webhook accepts Function creation")

		By("comparing the created Function with the current defaults")

		spec := function.FromUnstructured(fnUnstr).Spec
		defaults := function.Defaults(f.ClientSet)
		Expect(spec.Runtime).To(Equal(defaults.Runtime), "runtime")
		Expect(spec.Size).To(Equal(defaults.Size), "size")
		Expect(spec.Timeout).To(Equal(defaults.TimeOut), "timeout")
		Expect(spec.FunctionContentType).To(Equal(defaults.FuncContentType), "content type")
	})
})

func newServiceAccount(name string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
